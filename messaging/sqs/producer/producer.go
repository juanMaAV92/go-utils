package producer

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/juanMaAV92/go-utils/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	maxBatchSize         = 10
	maxMessageAttributes = 10
	warnAttributeLimit   = 8
)

// sqsAPI is the subset of *sqs.Client used by producer — enables mocking in tests.
type sqsAPI interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	SendMessageBatch(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)
}

type producer struct {
	client sqsAPI
	logger logger.Logger
	cfg    ProducerConfig
	name   string
	tracer trace.Tracer
}

// New creates a new SQS producer.
// name identifies this producer in logs and traces (useful when sending to multiple queues).
func New(client *sqs.Client, log logger.Logger, cfg ProducerConfig, name string) (Producer, error) {
	if client == nil {
		return nil, errors.New("sqs/producer: client is required")
	}
	if log == nil {
		return nil, errors.New("sqs/producer: logger is required")
	}
	if cfg.QueueURL == "" {
		return nil, errors.New("sqs/producer: queue URL is required")
	}
	if name == "" {
		return nil, errors.New("sqs/producer: name is required")
	}
	return newWithAPI(client, log, cfg, name), nil
}

// newWithAPI allows injecting a mock sqsAPI for tests.
func newWithAPI(client sqsAPI, log logger.Logger, cfg ProducerConfig, name string) Producer {
	return &producer{
		client: client,
		logger: log,
		cfg:    cfg,
		name:   name,
		tracer: otel.Tracer("github.com/juanMaAV92/go-utils/messaging/sqs"),
	}
}

// SendMessage sends a single message to SQS.
func (p *producer) SendMessage(ctx context.Context, msg *Message) error {
	if msg == nil {
		return errors.New("sqs/producer: message is required")
	}
	if msg.Body == "" {
		return errors.New("sqs/producer: message body cannot be empty")
	}

	ctx, span := p.tracer.Start(ctx,
		fmt.Sprintf("SQS send %s", p.name),
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(messagingAttrs(p.cfg.QueueURL, p.name)...),
	)
	defer span.End()
	span.SetAttributes(attribute.Int("messaging.message.body.size", len(msg.Body)))

	attrs, err := p.buildAttrsWithTrace(ctx, msg.Attributes)
	if err != nil {
		return p.fail(span, err)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:          aws.String(p.cfg.QueueURL),
		MessageBody:       aws.String(msg.Body),
		MessageAttributes: attrs,
	}
	if msg.MessageGroupId != "" {
		input.MessageGroupId = aws.String(msg.MessageGroupId)
	}
	if msg.MessageDeduplicationId != "" {
		input.MessageDeduplicationId = aws.String(msg.MessageDeduplicationId)
	}

	out, err := p.client.SendMessage(ctx, input)
	if err != nil {
		p.logError(ctx, "sqs.send_message", "failed to send message", "producer", p.name, "error", err.Error())
		return p.fail(span, fmt.Errorf("sqs: send message: %w", err))
	}

	msgID := aws.ToString(out.MessageId)
	span.SetAttributes(attribute.String("messaging.message.id", msgID))
	p.logInfo(ctx, "sqs.send_message", "message sent", "producer", p.name, "message_id", msgID)
	span.SetStatus(codes.Ok, "")
	return nil
}

// SendBatch sends up to 10 messages in a single SQS request.
func (p *producer) SendBatch(ctx context.Context, msgs []*Message) (*BatchResult, error) {
	if len(msgs) == 0 {
		return nil, errors.New("sqs/producer: at least one message is required")
	}
	if len(msgs) > maxBatchSize {
		return nil, fmt.Errorf("sqs/producer: batch size %d exceeds maximum of %d", len(msgs), maxBatchSize)
	}

	ctx, span := p.tracer.Start(ctx,
		fmt.Sprintf("SQS send_batch %s", p.name),
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(messagingAttrs(p.cfg.QueueURL, p.name)...),
	)
	defer span.End()

	entries := make([]types.SendMessageBatchRequestEntry, 0, len(msgs))
	for i, msg := range msgs {
		if msg == nil || msg.Body == "" {
			return nil, fmt.Errorf("sqs/producer: message at index %d has empty body", i)
		}
		attrs, err := p.buildAttrsWithTrace(ctx, msg.Attributes)
		if err != nil {
			return nil, p.fail(span, fmt.Errorf("message at index %d: %w", i, err))
		}
		entry := types.SendMessageBatchRequestEntry{
			Id:                aws.String(fmt.Sprintf("msg-%d", i)),
			MessageBody:       aws.String(msg.Body),
			MessageAttributes: attrs,
		}
		if msg.MessageGroupId != "" {
			entry.MessageGroupId = aws.String(msg.MessageGroupId)
		}
		if msg.MessageDeduplicationId != "" {
			entry.MessageDeduplicationId = aws.String(msg.MessageDeduplicationId)
		}
		entries = append(entries, entry)
	}

	out, err := p.client.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(p.cfg.QueueURL),
		Entries:  entries,
	})
	if err != nil {
		p.logError(ctx, "sqs.send_batch", "batch send failed", "producer", p.name, "error", err.Error())
		return nil, p.fail(span, fmt.Errorf("sqs: send batch: %w", err))
	}

	result := &BatchResult{
		SuccessCount: len(out.Successful),
		FailedCount:  len(out.Failed),
		FailedIds:    make([]string, 0, len(out.Failed)),
	}
	for _, f := range out.Failed {
		result.FailedIds = append(result.FailedIds, aws.ToString(f.Id))
	}

	span.SetAttributes(
		attribute.Int("messaging.batch.message_count", len(msgs)),
		attribute.Int("messaging.batch.success_count", result.SuccessCount),
		attribute.Int("messaging.batch.failed_count", result.FailedCount),
	)

	if result.FailedCount > 0 {
		p.logWarning(ctx, "sqs.send_batch", "partial batch failure",
			"producer", p.name, "success", result.SuccessCount, "failed", result.FailedCount)
		return result, fmt.Errorf("sqs: %d of %d messages failed", result.FailedCount, len(msgs))
	}

	p.logInfo(ctx, "sqs.send_batch", "batch sent", "producer", p.name, "count", result.SuccessCount)
	span.SetStatus(codes.Ok, "")
	return result, nil
}

// --- helpers ---

func (p *producer) logInfo(ctx context.Context, step, msg string, kv ...any) {
	if p.logger != nil {
		p.logger.Info(ctx, step, msg, kv...)
	}
}

func (p *producer) logWarning(ctx context.Context, step, msg string, kv ...any) {
	if p.logger != nil {
		p.logger.Warning(ctx, step, msg, kv...)
	}
}

func (p *producer) logError(ctx context.Context, step, msg string, kv ...any) {
	if p.logger != nil {
		p.logger.Error(ctx, step, msg, kv...)
	}
}

func (p *producer) fail(span trace.Span, err error) error {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}

// buildAttrsWithTrace converts user attributes to SQS format and injects W3C trace context.
// Returns an error if the total would exceed the AWS SQS limit of 10 attributes.
func (p *producer) buildAttrsWithTrace(ctx context.Context, userAttrs map[string]string) (map[string]types.MessageAttributeValue, error) {
	result := make(map[string]types.MessageAttributeValue, len(userAttrs)+2)
	for k, v := range userAttrs {
		result[k] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(v),
		}
	}

	carrier := &attributeCarrier{attrs: make(map[string]string)}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, v := range carrier.attrs {
		result[k] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(v),
		}
	}

	if len(result) > maxMessageAttributes {
		return nil, fmt.Errorf("sqs/producer: %d attributes exceed AWS limit of %d (user: %d, trace: %d)",
			len(result), maxMessageAttributes, len(userAttrs), len(carrier.attrs))
	}
	if len(userAttrs) >= warnAttributeLimit {
		p.logWarning(ctx, "sqs.validate_attrs", "approaching SQS attribute limit",
			"user_attrs", len(userAttrs), "trace_attrs", len(carrier.attrs), "total", len(result))
	}
	return result, nil
}

func messagingAttrs(queueURL, name string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("messaging.system", "sqs"),
		attribute.String("messaging.operation", "publish"),
		attribute.String("messaging.destination.name", queueURL),
		attribute.String("messaging.destination.kind", "queue"),
		attribute.String("cloud.provider", "aws"),
		attribute.String("messaging.producer.name", name),
	}
}

// attributeCarrier implements propagation.TextMapCarrier for SQS message attributes.
type attributeCarrier struct {
	attrs map[string]string
}

func (c *attributeCarrier) Get(key string) string        { return c.attrs[key] }
func (c *attributeCarrier) Set(key, value string)        { c.attrs[key] = value }
func (c *attributeCarrier) Keys() []string {
	keys := make([]string, 0, len(c.attrs))
	for k := range c.attrs {
		keys = append(keys, k)
	}
	return keys
}
