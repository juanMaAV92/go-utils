package consumer

import (
	"context"
	"fmt"

	"github.com/juanmaAV/go-utils/env"
)

// ConsumerConfig holds consumer-specific configuration.
type ConsumerConfig struct {
	QueueURL          string
	MaxMessages       int32 // 1–10, default 10
	WaitTimeSeconds   int32 // 0–20 (long polling), default 20
	VisibilityTimeout int32 // seconds, default 30
	WorkerPoolSize    int   // concurrent workers, default 10
}

// ConfigFromEnv reads consumer configuration from environment variables.
//
//	ConfigFromEnv("SQS")       → SQS_QUEUE_URL, SQS_MAX_MESSAGES, …
//	ConfigFromEnv("ORDER_SQS") → ORDER_SQS_QUEUE_URL, …
//
// Required: {prefix}_QUEUE_URL
func ConfigFromEnv(prefix string) (ConsumerConfig, error) {
	p := prefix + "_"
	cfg := ConsumerConfig{
		QueueURL:          env.GetEnv(p + "QUEUE_URL"),
		MaxMessages:       int32(env.GetEnvAsIntWithDefault(p+"MAX_MESSAGES", 10)),
		WaitTimeSeconds:   int32(env.GetEnvAsIntWithDefault(p+"WAIT_TIME_SECONDS", 20)),
		VisibilityTimeout: int32(env.GetEnvAsIntWithDefault(p+"VISIBILITY_TIMEOUT", 30)),
		WorkerPoolSize:    env.GetEnvAsIntWithDefault(p+"WORKER_POOL_SIZE", 10),
	}
	if cfg.QueueURL == "" {
		return ConsumerConfig{}, fmt.Errorf("sqs/consumer: missing required env var: %s", p+"QUEUE_URL")
	}
	return cfg, nil
}

// MessageProcessor is implemented by services to handle individual messages.
// Return nil to delete the message from the queue.
// Return an error to leave the message for retry (visibility timeout applies).
type MessageProcessor interface {
	ProcessMessage(ctx context.Context, body []byte) error
}

// Consumer is the interface for consuming messages from SQS.
type Consumer interface {
	// Start polls SQS and dispatches messages to the worker pool.
	// Blocks until ctx is cancelled.
	Start(ctx context.Context) error
}
