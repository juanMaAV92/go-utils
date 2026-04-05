package producer

import (
	"context"
	"fmt"

	"github.com/juanmaAV/go-utils/env"
)

// ProducerConfig holds producer-specific configuration.
type ProducerConfig struct {
	TopicArn string
}

// ConfigFromEnv reads producer configuration from environment variables.
//
//	ConfigFromEnv("SNS")       → SNS_TOPIC_ARN
//	ConfigFromEnv("ORDER_SNS") → ORDER_SNS_TOPIC_ARN
func ConfigFromEnv(prefix string) (ProducerConfig, error) {
	p := prefix + "_"
	cfg := ProducerConfig{
		TopicArn: env.GetEnv(p + "TOPIC_ARN"),
	}
	if cfg.TopicArn == "" {
		return ProducerConfig{}, fmt.Errorf("sns/producer: missing required env var: %s", p+"TOPIC_ARN")
	}
	return cfg, nil
}

// Message is a message to be published to SNS.
type Message struct {
	Body       string            // required
	Attributes map[string]string // optional (max 8 — 2 reserved for trace context)
	Subject    string            // optional; used by some subscription protocols (e.g. email)

	// FIFO topics only — ignored for standard topics.
	MessageGroupId         string
	MessageDeduplicationId string
}

// Producer is the interface for publishing messages to SNS.
type Producer interface {
	// Publish sends a single message to the configured topic.
	Publish(ctx context.Context, msg *Message) error
}
