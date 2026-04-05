package producer

import (
	"context"
	"fmt"

	"github.com/juanmaAV/go-utils/env"
)

// ProducerConfig holds producer-specific configuration.
type ProducerConfig struct {
	QueueURL string
}

// ConfigFromEnv reads producer configuration from environment variables.
//
//	ConfigFromEnv("SQS")       → SQS_QUEUE_URL
//	ConfigFromEnv("ORDER_SQS") → ORDER_SQS_QUEUE_URL
func ConfigFromEnv(prefix string) (ProducerConfig, error) {
	p := prefix + "_"
	cfg := ProducerConfig{
		QueueURL: env.GetEnv(p + "QUEUE_URL"),
	}
	if cfg.QueueURL == "" {
		return ProducerConfig{}, fmt.Errorf("sqs/producer: missing required env var: %s", p+"QUEUE_URL")
	}
	return cfg, nil
}

// Message is a message to be sent to SQS.
type Message struct {
	Body       string            // required
	Attributes map[string]string // optional custom attributes (max 8 — 2 reserved for trace context)

	// FIFO queues only — ignored for standard queues.
	MessageGroupId         string
	MessageDeduplicationId string
}

// BatchResult contains the outcome of a SendBatch call.
type BatchResult struct {
	SuccessCount int
	FailedCount  int
	FailedIds    []string // batch entry IDs that failed
}

// Producer is the interface for sending messages to SQS.
type Producer interface {
	// SendMessage sends a single message.
	SendMessage(ctx context.Context, msg *Message) error

	// SendBatch sends up to 10 messages in a single request.
	// Returns BatchResult even on partial failure.
	SendBatch(ctx context.Context, msgs []*Message) (*BatchResult, error)
}
