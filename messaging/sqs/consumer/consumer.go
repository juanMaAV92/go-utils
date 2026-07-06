package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/juanMaAV92/go-utils/v2/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const messageChanBuffer = 100

// sqsAPI is the subset of *sqs.Client used by consumer — enables mocking in tests.
type sqsAPI interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

// snsEnvelope is the wrapper SNS adds when delivering to SQS.
type snsEnvelope struct {
	Type              string                    `json:"Type"`
	TopicArn          string                    `json:"TopicArn"`
	Message           string                    `json:"Message"`
	MessageAttributes map[string]snsMessageAttr `json:"MessageAttributes,omitempty"`
}

type snsMessageAttr struct {
	Type  string `json:"Type"`
	Value string `json:"Value"`
}

type consumer struct {
	client      sqsAPI
	processor   MessageProcessor
	logger      logger.Logger
	cfg         ConsumerConfig
	name        string
	messageChan chan types.Message
	wg          sync.WaitGroup
	tracer      trace.Tracer
}

// New creates a new SQS consumer.
// name identifies this consumer in logs and traces.
func New(client *sqs.Client, processor MessageProcessor, log logger.Logger, cfg ConsumerConfig, name string) (Consumer, error) {
	if client == nil {
		return nil, errors.New("sqs/consumer: client is required")
	}
	if processor == nil {
		return nil, errors.New("sqs/consumer: processor is required")
	}
	if log == nil {
		return nil, errors.New("sqs/consumer: logger is required")
	}
	if cfg.QueueURL == "" {
		return nil, errors.New("sqs/consumer: queue URL is required")
	}
	if name == "" {
		return nil, errors.New("sqs/consumer: name is required")
	}
	return newWithAPI(client, processor, log, cfg, name), nil
}

func newWithAPI(client sqsAPI, processor MessageProcessor, log logger.Logger, cfg ConsumerConfig, name string) Consumer {
	cfg = withConsumerDefaults(cfg)
	return &consumer{
		client:      client,
		processor:   processor,
		logger:      log,
		cfg:         cfg,
		name:        name,
		messageChan: make(chan types.Message, messageChanBuffer),
		tracer:      otel.Tracer("github.com/juanMaAV92/go-utils/v2/messaging/sqs"),
	}
}

// withConsumerDefaults fills zero-valued fields and clamps MaxMessages /
// WaitTimeSeconds to the AWS limits (1–10 and 0–20). Sending out-of-range
// values makes ReceiveMessage fail on every call.
func withConsumerDefaults(cfg ConsumerConfig) ConsumerConfig {
	if cfg.MaxMessages < 1 {
		cfg.MaxMessages = 10
	} else if cfg.MaxMessages > 10 {
		cfg.MaxMessages = 10
	}
	if cfg.WaitTimeSeconds < 0 {
		cfg.WaitTimeSeconds = 0
	} else if cfg.WaitTimeSeconds > 20 {
		cfg.WaitTimeSeconds = 20
	}
	if cfg.VisibilityTimeout <= 0 {
		cfg.VisibilityTimeout = 30
	}
	if cfg.WorkerPoolSize <= 0 {
		cfg.WorkerPoolSize = 10
	}
	if cfg.PollErrorBackoff <= 0 {
		cfg.PollErrorBackoff = time.Second
	}
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 30 * time.Second
	}
	return cfg
}

// Start polls SQS and dispatches messages to the worker pool.
// Blocks until ctx is cancelled.
func (c *consumer) Start(ctx context.Context) error {
	c.logger.Info(ctx, "sqs.consumer.start", "starting consumer",
		"consumer", c.name, "queue", c.cfg.QueueURL, "workers", c.cfg.WorkerPoolSize)

	// Workers run on a context that survives ctx cancellation, so in-flight
	// processing and its subsequent delete complete cleanly on shutdown instead
	// of failing on a cancelled ctx (which would redeliver already-handled
	// messages). It is force-cancelled only if the drain exceeds ShutdownTimeout.
	workerCtx, cancelWorkers := context.WithCancel(context.WithoutCancel(ctx))
	defer cancelWorkers()

	for i := 0; i < c.cfg.WorkerPoolSize; i++ {
		c.wg.Add(1)
		go c.worker(workerCtx, i)
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Info(ctx, "sqs.consumer.stop", "context cancelled, draining", "consumer", c.name)
			close(c.messageChan)
			c.drain(ctx, cancelWorkers)
			c.logger.Info(ctx, "sqs.consumer.stop", "consumer stopped", "consumer", c.name)
			return ctx.Err()
		default:
			if err := c.poll(ctx); err != nil {
				c.logger.Error(ctx, "sqs.consumer.poll", "poll error, backing off",
					"consumer", c.name, "backoff", c.cfg.PollErrorBackoff.String(), "error", err.Error())
				// Backoff prevents a tight hot loop that hammers SQS (and amplifies
				// throttling) when ReceiveMessage keeps failing.
				select {
				case <-time.After(c.cfg.PollErrorBackoff):
				case <-ctx.Done():
				}
			}
		}
	}
}

// drain waits for workers to finish buffered messages, bounded by ShutdownTimeout.
func (c *consumer) drain(ctx context.Context, cancelWorkers context.CancelFunc) {
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(c.cfg.ShutdownTimeout):
		c.logger.Warning(ctx, "sqs.consumer.stop", "drain timed out, cancelling in-flight work",
			"consumer", c.name, "timeout", c.cfg.ShutdownTimeout.String())
		cancelWorkers()
		<-done
	}
}

func (c *consumer) worker(ctx context.Context, workerID int) {
	defer c.wg.Done()
	for msg := range c.messageChan {
		c.process(ctx, msg, workerID)
	}
}

func (c *consumer) poll(ctx context.Context) error {
	out, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(c.cfg.QueueURL),
		MaxNumberOfMessages:   c.cfg.MaxMessages,
		WaitTimeSeconds:       c.cfg.WaitTimeSeconds,
		VisibilityTimeout:     c.cfg.VisibilityTimeout,
		MessageAttributeNames: []string{"All"}, // required to receive trace context
	})
	if err != nil {
		return fmt.Errorf("sqs: receive message: %w", err)
	}
	for _, msg := range out.Messages {
		select {
		case c.messageChan <- msg:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (c *consumer) process(ctx context.Context, msg types.Message, workerID int) {
	msgID := aws.ToString(msg.MessageId)
	body := aws.ToString(msg.Body)
	attrs := msg.MessageAttributes

	// Unwrap SNS envelope if present — allows the same consumer to handle
	// direct SQS messages and SNS-proxied messages transparently.
	if env, ok := unwrapSNS(body); ok {
		body = env.Message
		if attrs == nil {
			attrs = make(map[string]types.MessageAttributeValue)
		}
		for k, v := range env.MessageAttributes {
			attrs[k] = types.MessageAttributeValue{
				DataType:    aws.String(v.Type),
				StringValue: aws.String(v.Value),
			}
		}
	}

	// Restore distributed trace from message attributes.
	ctx = extractTrace(ctx, attrs)

	ctx, span := c.tracer.Start(ctx,
		fmt.Sprintf("SQS process %s", c.name),
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "sqs"),
			attribute.String("messaging.operation", "process"),
			attribute.String("messaging.destination.name", c.cfg.QueueURL),
			attribute.String("messaging.destination.kind", "queue"),
			attribute.String("cloud.provider", "aws"),
			attribute.String("messaging.consumer.name", c.name),
			attribute.String("messaging.message.id", msgID),
			attribute.Int("messaging.worker.id", workerID),
			attribute.Int("messaging.message.body.size", len(body)),
		),
	)
	defer span.End()

	if err := c.processor.ProcessMessage(ctx, []byte(body)); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		c.logger.Error(ctx, "sqs.consumer.process", "processing failed, message will retry",
			"consumer", c.name, "worker", workerID, "message_id", msgID, "error", err.Error())
		return // do not delete — SQS will redeliver after visibility timeout
	}

	if err := c.deleteMessage(ctx, aws.ToString(msg.ReceiptHandle)); err != nil {
		span.RecordError(err)
		c.logger.Error(ctx, "sqs.consumer.delete", "failed to delete message",
			"consumer", c.name, "worker", workerID, "message_id", msgID, "error", err.Error())
		return
	}

	span.SetStatus(codes.Ok, "")
	c.logger.Info(ctx, "sqs.consumer.process", "message processed",
		"consumer", c.name, "worker", workerID, "message_id", msgID)
}

func (c *consumer) deleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.cfg.QueueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("sqs: delete message: %w", err)
	}
	return nil
}

// unwrapSNS detects and parses SNS envelope payloads.
func unwrapSNS(body string) (*snsEnvelope, bool) {
	if !strings.Contains(body, `"Type"`) || !strings.Contains(body, `"TopicArn"`) {
		return nil, false
	}
	var env snsEnvelope
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		return nil, false
	}
	if env.Type != "Notification" || env.TopicArn == "" || env.Message == "" {
		return nil, false
	}
	return &env, true
}

// extractTrace restores W3C Trace Context from SQS message attributes into ctx.
func extractTrace(ctx context.Context, attrs map[string]types.MessageAttributeValue) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	carrier := &attributeCarrier{attrs: make(map[string]string, len(attrs))}
	for k, v := range attrs {
		if v.StringValue != nil {
			carrier.attrs[k] = *v.StringValue
		}
	}
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

// attributeCarrier implements propagation.TextMapCarrier for SQS message attributes.
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
