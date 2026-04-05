package sns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/juanmaAV/go-utils/logger"
)

// NewClient creates an AWS SNS client.
// The returned *sns.Client is passed to producer.New.
func NewClient(ctx context.Context, cfg Config, log logger.Logger) (*sns.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("sns: failed to load AWS config: %w", err)
	}

	opts := []func(*sns.Options){}
	if cfg.Endpoint != "" {
		opts = append(opts, func(o *sns.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
		if log != nil {
			log.Info(ctx, "sns.new_client", "using custom SNS endpoint", "endpoint", cfg.Endpoint)
		}
	}

	client := sns.NewFromConfig(awsCfg, opts...)
	if log != nil {
		log.Info(ctx, "sns.new_client", "SNS client created", "region", cfg.Region)
	}
	return client, nil
}
