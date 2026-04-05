package sqs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/juanmaAV/go-utils/logger"
)

// NewClient creates an AWS SQS client.
// The returned *sqs.Client is passed to producer.New or consumer.New.
func NewClient(ctx context.Context, cfg Config, log logger.Logger) (*sqs.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("sqs: failed to load AWS config: %w", err)
	}

	opts := []func(*sqs.Options){}
	if cfg.Endpoint != "" {
		opts = append(opts, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
		if log != nil {
			log.Info(ctx, "sqs.new_client", "using custom SQS endpoint", "endpoint", cfg.Endpoint)
		}
	}

	client := sqs.NewFromConfig(awsCfg, opts...)
	if log != nil {
		log.Info(ctx, "sqs.new_client", "SQS client created", "region", cfg.Region)
	}
	return client, nil
}
