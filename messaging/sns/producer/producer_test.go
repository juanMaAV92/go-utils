package producer

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"go.opentelemetry.io/otel"
)

// --- mock ---

type mockSNS struct {
	publishFn func(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

func (m *mockSNS) Publish(ctx context.Context, p *sns.PublishInput, o ...func(*sns.Options)) (*sns.PublishOutput, error) {
	return m.publishFn(ctx, p, o...)
}

func newTestProducer(mock snsAPI) Producer {
	return newWithAPI(mock, nil, ProducerConfig{TopicArn: "arn:aws:sns:us-east-1:123:my-topic"}, "test-producer")
}

var ctx = context.Background()

// ---- Publish ----

func TestPublish_Success(t *testing.T) {
	var capturedBody string
	p := newTestProducer(&mockSNS{
		publishFn: func(_ context.Context, params *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
			capturedBody = aws.ToString(params.Message)
			return &sns.PublishOutput{MessageId: aws.String("msg-1")}, nil
		},
	})

	if err := p.Publish(ctx, &Message{Body: "hello"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody != "hello" {
		t.Errorf("body = %q, want hello", capturedBody)
	}
}

func TestPublish_NilMessage(t *testing.T) {
	p := newTestProducer(&mockSNS{})
	if err := p.Publish(ctx, nil); err == nil {
		t.Error("expected error for nil message")
	}
}

func TestPublish_EmptyBody(t *testing.T) {
	p := newTestProducer(&mockSNS{})
	if err := p.Publish(ctx, &Message{Body: ""}); err == nil {
		t.Error("expected error for empty body")
	}
}

func TestPublish_Subject(t *testing.T) {
	var capturedSubject string
	p := newTestProducer(&mockSNS{
		publishFn: func(_ context.Context, params *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
			capturedSubject = aws.ToString(params.Subject)
			return &sns.PublishOutput{MessageId: aws.String("msg-1")}, nil
		},
	})

	if err := p.Publish(ctx, &Message{Body: "hello", Subject: "Order Placed"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedSubject != "Order Placed" {
		t.Errorf("subject = %q, want Order Placed", capturedSubject)
	}
}

func TestPublish_FIFO(t *testing.T) {
	var capturedGroupId, capturedDeduplicationId string
	p := newTestProducer(&mockSNS{
		publishFn: func(_ context.Context, params *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
			capturedGroupId = aws.ToString(params.MessageGroupId)
			capturedDeduplicationId = aws.ToString(params.MessageDeduplicationId)
			return &sns.PublishOutput{MessageId: aws.String("msg-fifo")}, nil
		},
	})

	err := p.Publish(ctx, &Message{
		Body:                   "event",
		MessageGroupId:         "order-group",
		MessageDeduplicationId: "dedup-456",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedGroupId != "order-group" {
		t.Errorf("MessageGroupId = %q, want order-group", capturedGroupId)
	}
	if capturedDeduplicationId != "dedup-456" {
		t.Errorf("MessageDeduplicationId = %q, want dedup-456", capturedDeduplicationId)
	}
}

func TestPublish_WithAttributes(t *testing.T) {
	var capturedAttrs map[string]string
	p := newTestProducer(&mockSNS{
		publishFn: func(_ context.Context, params *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
			capturedAttrs = make(map[string]string)
			for k, v := range params.MessageAttributes {
				capturedAttrs[k] = aws.ToString(v.StringValue)
			}
			return &sns.PublishOutput{MessageId: aws.String("msg-1")}, nil
		},
	})

	err := p.Publish(ctx, &Message{
		Body:       "event",
		Attributes: map[string]string{"source": "order-service", "version": "2"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedAttrs["source"] != "order-service" {
		t.Errorf("attribute source = %q, want order-service", capturedAttrs["source"])
	}
}

func TestPublish_TraceContextInjected(t *testing.T) {
	_ = otel.GetTextMapPropagator() // ensure no panic with noop propagator
	p := newTestProducer(&mockSNS{
		publishFn: func(_ context.Context, _ *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
			return &sns.PublishOutput{MessageId: aws.String("msg-1")}, nil
		},
	})
	if err := p.Publish(ctx, &Message{Body: "traced"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPublish_SDKError(t *testing.T) {
	p := newTestProducer(&mockSNS{
		publishFn: func(_ context.Context, _ *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
			return nil, errors.New("connection refused")
		},
	})
	if err := p.Publish(ctx, &Message{Body: "hello"}); err == nil {
		t.Error("expected error from SDK")
	}
}

func TestPublish_TooManyAttributes(t *testing.T) {
	// 11 user attributes always exceeds the limit of 10, regardless of trace context injection.
	p := newTestProducer(&mockSNS{})
	attrs := make(map[string]string, 11)
	for i := 0; i < 11; i++ {
		attrs[fmt.Sprintf("key%d", i)] = "val"
	}
	if err := p.Publish(ctx, &Message{Body: "hello", Attributes: attrs}); err == nil {
		t.Error("expected error when attributes exceed limit")
	}
}

// ---- New validation ----

func TestNew_NilClient(t *testing.T) {
	_, err := New(nil, nil, ProducerConfig{TopicArn: "arn"}, "name")
	if err == nil {
		t.Error("expected error for nil client")
	}
}
