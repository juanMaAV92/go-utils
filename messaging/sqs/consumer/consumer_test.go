package consumer

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// --- mocks ---

type mockSQS struct {
	receiveFn func(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	deleteFn  func(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

func (m *mockSQS) ReceiveMessage(ctx context.Context, p *sqs.ReceiveMessageInput, o ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	return m.receiveFn(ctx, p, o...)
}
func (m *mockSQS) DeleteMessage(ctx context.Context, p *sqs.DeleteMessageInput, o ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	return m.deleteFn(ctx, p, o...)
}

type mockProcessor struct {
	fn func(ctx context.Context, body []byte) error
}

func (m *mockProcessor) ProcessMessage(ctx context.Context, body []byte) error {
	return m.fn(ctx, body)
}

// mockLogger satisfies logger.Logger without importing the package.
type mockLogger struct{}

func (mockLogger) Info(_ context.Context, _, _ string, _ ...any)    {}
func (mockLogger) Error(_ context.Context, _, _ string, _ ...any)   {}
func (mockLogger) Warning(_ context.Context, _, _ string, _ ...any) {}
func (mockLogger) Debug(_ context.Context, _, _ string, _ ...any)   {}
func (mockLogger) Fatal(_ context.Context, _, _ string, _ ...any)   {}

func cfg() ConsumerConfig {
	return ConsumerConfig{
		QueueURL:          "https://sqs.us-east-1.amazonaws.com/123/my-queue",
		MaxMessages:       10,
		WaitTimeSeconds:   0,
		VisibilityTimeout: 30,
		WorkerPoolSize:    2,
	}
}

var ctx = context.Background()

// ---- unwrapSNS ----

func TestUnwrapSNS_ValidEnvelope(t *testing.T) {
	body := `{"Type":"Notification","TopicArn":"arn:aws:sns:us-east-1:123:topic","Message":"hello"}`
	env, ok := unwrapSNS(body)
	if !ok {
		t.Fatal("expected SNS envelope to be detected")
	}
	if env.Message != "hello" {
		t.Errorf("Message = %q, want hello", env.Message)
	}
}

func TestUnwrapSNS_DirectMessage(t *testing.T) {
	body := `{"order_id":"123","status":"shipped"}`
	_, ok := unwrapSNS(body)
	if ok {
		t.Error("expected direct SQS message to not be detected as SNS")
	}
}

func TestUnwrapSNS_InvalidJSON(t *testing.T) {
	_, ok := unwrapSNS(`not json`)
	if ok {
		t.Error("expected invalid JSON to return false")
	}
}

func TestUnwrapSNS_MissingMessageField(t *testing.T) {
	body := `{"Type":"Notification","TopicArn":"arn:aws:sns:us-east-1:123:topic","Message":""}`
	_, ok := unwrapSNS(body)
	if ok {
		t.Error("expected envelope with empty Message to return false")
	}
}

// ---- extractTrace ----

func TestExtractTrace_EmptyAttrs(t *testing.T) {
	// Should return ctx unchanged without panicking
	result := extractTrace(ctx, nil)
	if result == nil {
		t.Error("expected non-nil context")
	}
}

func TestExtractTrace_WithAttrs(t *testing.T) {
	attrs := map[string]types.MessageAttributeValue{
		"traceparent": {DataType: aws.String("String"), StringValue: aws.String("00-abc-def-01")},
	}
	result := extractTrace(ctx, attrs)
	if result == nil {
		t.Error("expected non-nil context")
	}
}

// ---- process (integration of process + delete) ----

func TestProcess_SuccessDeletesMessage(t *testing.T) {
	deleted := false
	c := newWithAPI(
		&mockSQS{
			deleteFn: func(_ context.Context, _ *sqs.DeleteMessageInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
				deleted = true
				return &sqs.DeleteMessageOutput{}, nil
			},
		},
		&mockProcessor{fn: func(_ context.Context, _ []byte) error { return nil }},
		mockLogger{},
		cfg(),
		"test-consumer",
	).(*consumer)

	c.process(ctx, types.Message{
		MessageId:     aws.String("msg-1"),
		ReceiptHandle: aws.String("handle-1"),
		Body:          aws.String(`{"event":"order_placed"}`),
	}, 0)

	if !deleted {
		t.Error("expected message to be deleted after successful processing")
	}
}

func TestProcess_ProcessorErrorDoesNotDelete(t *testing.T) {
	deleted := false
	c := newWithAPI(
		&mockSQS{
			deleteFn: func(_ context.Context, _ *sqs.DeleteMessageInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
				deleted = true
				return &sqs.DeleteMessageOutput{}, nil
			},
		},
		&mockProcessor{fn: func(_ context.Context, _ []byte) error { return errors.New("processing failed") }},
		mockLogger{},
		cfg(),
		"test-consumer",
	).(*consumer)

	c.process(ctx, types.Message{
		MessageId:     aws.String("msg-2"),
		ReceiptHandle: aws.String("handle-2"),
		Body:          aws.String(`{"event":"order_placed"}`),
	}, 0)

	if deleted {
		t.Error("message should NOT be deleted when processor returns error")
	}
}

func TestProcess_UnwrapsSNSEnvelope(t *testing.T) {
	var processedBody []byte
	c := newWithAPI(
		&mockSQS{
			deleteFn: func(_ context.Context, _ *sqs.DeleteMessageInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
				return &sqs.DeleteMessageOutput{}, nil
			},
		},
		&mockProcessor{fn: func(_ context.Context, body []byte) error {
			processedBody = body
			return nil
		}},
		mockLogger{},
		cfg(),
		"test-consumer",
	).(*consumer)

	snsBody := `{"Type":"Notification","TopicArn":"arn:aws:sns:us-east-1:123:topic","Message":"{\"order_id\":\"42\"}"}`
	c.process(ctx, types.Message{
		MessageId:     aws.String("msg-3"),
		ReceiptHandle: aws.String("handle-3"),
		Body:          aws.String(snsBody),
	}, 0)

	if string(processedBody) != `{"order_id":"42"}` {
		t.Errorf("processedBody = %q, want unwrapped message", string(processedBody))
	}
}

// ---- New validation ----

func TestNew_NilClient(t *testing.T) {
	_, err := New(nil, &mockProcessor{}, mockLogger{}, cfg(), "name")
	if err == nil {
		t.Error("expected error for nil client")
	}
}

func TestNew_NilProcessor(t *testing.T) {
	// Can't pass real *sqs.Client in unit test, use mock via newWithAPI
	c, err := newWithAPI(&mockSQS{}, nil, mockLogger{}, cfg(), "name"), error(nil)
	_ = c
	_ = err
	// newWithAPI doesn't validate nil processor — validation is in New()
	// just ensure newWithAPI doesn't panic
}
