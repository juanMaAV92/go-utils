package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/juanMaAV92/go-utils/logger"
	"go.opentelemetry.io/otel"
)

// New creates a Storage client backed by AWS S3.
// Call once at service startup and inject the returned Storage where needed.
func New(ctx context.Context, cfg Config, log logger.Logger) (Storage, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("s3: failed to load AWS config: %w", err)
	}

	opts := []func(*s3.Options){}
	if cfg.Endpoint != "" {
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true // required for LocalStack
		})
	}

	client := s3.NewFromConfig(awsCfg, opts...)

	return &storage{
		client:    client,
		presigner: s3.NewPresignClient(client),
		logger:    log,
		tracer:    otel.Tracer("github.com/juanMaAV92/go-utils/storage/s3"),
	}, nil
}
