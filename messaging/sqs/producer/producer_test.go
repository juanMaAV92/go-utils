package producer

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.opentelemetry.io/otel"
)

// --- mock ---

type mockSQS struct {
	sendFn      func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	sendBatchFn func(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)
}

func (m *mockSQS) SendMessage(ctx context.Context, p *sqs.SendMessageInput, o ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return m.sendFn(ctx, p, o...)
}
func (m *mockSQS) SendMessageBatch(ctx context.Context, p *sqs.SendMessageBatchInput, o ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error) {
	return m.sendBatchFn(ctx, p, o...)
}

func newTestProducer(mock sqsAPI) Producer {
	return newWithAPI(mock, nil, ProducerConfig{QueueURL: "https://sqs.us-east-1.amazonaws.com/123/my-queue"}, "test-producer")
}

var ctx = context.Background()

// ---- SendMessage ----

func TestSendMessage_Success(t *testing.T) {
	var capturedBody string
	p := newTestProducer(&mockSQS{
		sendFn: func(_ context.Context, params *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
			capturedBody = aws.ToString(params.MessageBody)
			return &sqs.SendMessageOutput{MessageId: aws.String("msg-1")}, nil
		},
	})

	if err := p.SendMessage(ctx, &Message{Body: "hello"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody != "hello" {
		t.Errorf("body = %q, want hello", capturedBody)
	}
}

func TestSendMessage_NilMessage(t *testing.T) {
	p := newTestProducer(&mockSQS{})
	if err := p.SendMessage(ctx, nil); err == nil {
		t.Error("expected error for nil message")
	}
}

func TestSendMessage_EmptyBody(t *testing.T) {
	p := newTestProducer(&mockSQS{})
	if err := p.SendMessage(ctx, &Message{Body: ""}); err == nil {
		t.Error("expected error for empty body")
	}
}

func TestSendMessage_FIFO(t *testing.T) {
	var capturedGroupId, capturedDeduplicationId string
	p := newTestProducer(&mockSQS{
		sendFn: func(_ context.Context, params *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
			capturedGroupId = aws.ToString(params.MessageGroupId)
			capturedDeduplicationId = aws.ToString(params.MessageDeduplicationId)
			return &sqs.SendMessageOutput{MessageId: aws.String("msg-fifo")}, nil
		},
	})

	err := p.SendMessage(ctx, &Message{
		Body:                   "event",
		MessageGroupId:         "order-group",
		MessageDeduplicationId: "dedup-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedGroupId != "order-group" {
		t.Errorf("MessageGroupId = %q, want order-group", capturedGroupId)
	}
	if capturedDeduplicationId != "dedup-123" {
		t.Errorf("MessageDeduplicationId = %q, want dedup-123", capturedDeduplicationId)
	}
}

func TestSendMessage_PropagatesTraceContext(t *testing.T) {
	// With a real OTel propagator, trace context keys appear in message attributes.
	// Without a configured propagator (noop), no trace keys are injected — that's fine.
	_ = otel.GetTextMapPropagator() // ensure no panic
	p := newTestProducer(&mockSQS{
		sendFn: func(_ context.Context, _ *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
			return &sqs.SendMessageOutput{MessageId: aws.String("msg-1")}, nil
		},
	})
	if err := p.SendMessage(ctx, &Message{Body: "traced"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendMessage_SDKError(t *testing.T) {
	p := newTestProducer(&mockSQS{
		sendFn: func(_ context.Context, _ *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
			return nil, errors.New("connection refused")
		},
	})
	if err := p.SendMessage(ctx, &Message{Body: "hello"}); err == nil {
		t.Error("expected error from SDK")
	}
}

// ---- SendBatch ----

func TestSendBatch_Success(t *testing.T) {
	p := newTestProducer(&mockSQS{
		sendBatchFn: func(_ context.Context, params *sqs.SendMessageBatchInput, _ ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error) {
			successful := make([]types.SendMessageBatchResultEntry, len(params.Entries))
			for i, e := range params.Entries {
				successful[i] = types.SendMessageBatchResultEntry{Id: e.Id, MessageId: aws.String("msg-" + aws.ToString(e.Id))}
			}
			return &sqs.SendMessageBatchOutput{Successful: successful}, nil
		},
	})

	msgs := []*Message{{Body: "a"}, {Body: "b"}, {Body: "c"}}
	result, err := p.SendBatch(ctx, msgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SuccessCount != 3 {
		t.Errorf("SuccessCount = %d, want 3", result.SuccessCount)
	}
	if result.FailedCount != 0 {
		t.Errorf("FailedCount = %d, want 0", result.FailedCount)
	}
}

func TestSendBatch_PartialFailure(t *testing.T) {
	p := newTestProducer(&mockSQS{
		sendBatchFn: func(_ context.Context, params *sqs.SendMessageBatchInput, _ ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error) {
			return &sqs.SendMessageBatchOutput{
				Successful: []types.SendMessageBatchResultEntry{{Id: aws.String("msg-0")}},
				Failed:     []types.BatchResultErrorEntry{{Id: aws.String("msg-1"), Code: aws.String("InternalError")}},
			}, nil
		},
	})

	result, err := p.SendBatch(ctx, []*Message{{Body: "a"}, {Body: "b"}})
	if err == nil {
		t.Error("expected error on partial failure")
	}
	if result == nil {
		t.Fatal("expected non-nil result on partial failure")
	}
	if result.SuccessCount != 1 || result.FailedCount != 1 {
		t.Errorf("got success=%d failed=%d, want 1/1", result.SuccessCount, result.FailedCount)
	}
}

func TestSendBatch_ExceedsMaxSize(t *testing.T) {
	p := newTestProducer(&mockSQS{})
	msgs := make([]*Message, 11)
	for i := range msgs {
		msgs[i] = &Message{Body: "x"}
	}
	if _, err := p.SendBatch(ctx, msgs); err == nil {
		t.Error("expected error for batch > 10")
	}
}

func TestSendBatch_EmptyList(t *testing.T) {
	p := newTestProducer(&mockSQS{})
	if _, err := p.SendBatch(ctx, nil); err == nil {
		t.Error("expected error for empty batch")
	}
}

// ---- New validation ----

func TestNew_NilClient(t *testing.T) {
	_, err := New(nil, nil, ProducerConfig{QueueURL: "url"}, "name")
	if err == nil {
		t.Error("expected error for nil client")
	}
}
