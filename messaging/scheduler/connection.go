package scheduler

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/juanMaAV92/go-utils/logger"
	"go.opentelemetry.io/otel"
)

// New creates a Scheduler client backed by AWS EventBridge Scheduler.
func New(ctx context.Context, cfg Config, log logger.Logger) (Scheduler, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("scheduler: failed to load AWS config: %w", err)
	}

	opts := []func(*scheduler.Options){}
	if cfg.Endpoint != "" {
		opts = append(opts, func(o *scheduler.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	return &sched{
		client:  scheduler.NewFromConfig(awsCfg, opts...),
		logger:  log,
		tracer:  otel.Tracer("github.com/juanMaAV92/go-utils/messaging/scheduler"),
		roleArn: cfg.RoleArn,
	}, nil
}
