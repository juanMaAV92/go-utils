package producer

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/juanMaAV92/go-utils/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	maxMessageAttributes = 10
	warnAttributeLimit   = 8
)

// snsAPI is the subset of *sns.Client used by producer — enables mocking in tests.
type snsAPI interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

type producer struct {
	client snsAPI
	logger logger.Logger
	cfg    ProducerConfig
	name   string
	tracer trace.Tracer
}

// New creates a new SNS producer.
// name identifies this producer in logs and traces.
func New(client *sns.Client, log logger.Logger, cfg ProducerConfig, name string) (Producer, error) {
	if client == nil {
		return nil, errors.New("sns/producer: client is required")
	}
	if log == nil {
		return nil, errors.New("sns/producer: logger is required")
	}
	if cfg.TopicArn == "" {
		return nil, errors.New("sns/producer: topic ARN is required")
	}
	if name == "" {
		return nil, errors.New("sns/producer: name is required")
	}
	return newWithAPI(client, log, cfg, name), nil
}

// newWithAPI allows injecting a mock snsAPI for tests.
func newWithAPI(client snsAPI, log logger.Logger, cfg ProducerConfig, name string) Producer {
	return &producer{
		client: client,
		logger: log,
		cfg:    cfg,
		name:   name,
		tracer: otel.Tracer("github.com/juanMaAV92/go-utils/messaging/sns"),
	}
}

// Publish sends a message to the configured SNS topic.
func (p *producer) Publish(ctx context.Context, msg *Message) error {
	if msg == nil {
		return errors.New("sns/producer: message is required")
	}
	if msg.Body == "" {
		return errors.New("sns/producer: message body cannot be empty")
	}

	ctx, span := p.tracer.Start(ctx,
		fmt.Sprintf("SNS publish %s", p.name),
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "sns"),
			attribute.String("messaging.operation", "publish"),
			attribute.String("messaging.destination.name", p.cfg.TopicArn),
			attribute.String("messaging.destination.kind", "topic"),
			attribute.String("cloud.provider", "aws"),
			attribute.String("messaging.producer.name", p.name),
			attribute.Int("messaging.message.body.size", len(msg.Body)),
		),
	)
	defer span.End()

	attrs, err := p.buildAttrsWithTrace(ctx, msg.Attributes)
	if err != nil {
		return p.fail(span, err)
	}

	input := &sns.PublishInput{
		TopicArn:          aws.String(p.cfg.TopicArn),
		Message:           aws.String(msg.Body),
		MessageAttributes: attrs,
	}
	if msg.Subject != "" {
		input.Subject = aws.String(msg.Subject)
	}
	if msg.MessageGroupId != "" {
		input.MessageGroupId = aws.String(msg.MessageGroupId)
	}
	if msg.MessageDeduplicationId != "" {
		input.MessageDeduplicationId = aws.String(msg.MessageDeduplicationId)
	}

	out, err := p.client.Publish(ctx, input)
	if err != nil {
		p.logError(ctx, "sns.publish", "failed to publish message", "producer", p.name, "error", err.Error())
		return p.fail(span, fmt.Errorf("sns: publish: %w", err))
	}

	msgID := aws.ToString(out.MessageId)
	span.SetAttributes(attribute.String("messaging.message.id", msgID))
	p.logInfo(ctx, "sns.publish", "message published", "producer", p.name, "message_id", msgID)
	span.SetStatus(codes.Ok, "")
	return nil
}

// --- helpers ---

func (p *producer) logInfo(ctx context.Context, step, msg string, kv ...any) {
	if p.logger != nil {
		p.logger.Info(ctx, step, msg, kv...)
	}
}

func (p *producer) logError(ctx context.Context, step, msg string, kv ...any) {
	if p.logger != nil {
		p.logger.Error(ctx, step, msg, kv...)
	}
}

func (p *producer) logWarning(ctx context.Context, step, msg string, kv ...any) {
	if p.logger != nil {
		p.logger.Warning(ctx, step, msg, kv...)
	}
}

func (p *producer) fail(span trace.Span, err error) error {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}

// buildAttrsWithTrace converts user attributes to SNS format and injects W3C trace context.
// Returns an error if the total would exceed the AWS SNS limit of 10 attributes.
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
		return nil, fmt.Errorf("sns/producer: %d attributes exceed AWS limit of %d (user: %d, trace: %d)",
			len(result), maxMessageAttributes, len(userAttrs), len(carrier.attrs))
	}
	if len(userAttrs) >= warnAttributeLimit {
		p.logWarning(ctx, "sns.validate_attrs", "approaching SNS attribute limit",
			"user_attrs", len(userAttrs), "trace_attrs", len(carrier.attrs), "total", len(result))
	}
	return result, nil
}

// attributeCarrier implements propagation.TextMapCarrier for SNS message attributes.
type attributeCarrier struct {
	attrs map[string]string
}

func (c *attributeCarrier) Get(key string) string { return c.attrs[key] }
func (c *attributeCarrier) Set(key, value string) { c.attrs[key] = value }
func (c *attributeCarrier) Keys() []string {
	keys := make([]string, 0, len(c.attrs))
	for k := range c.attrs {
		keys = append(keys, k)
	}
	return keys
}
