package s3

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/juanMaAV92/go-utils/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultPresignExpiration = 5 * time.Minute
	maxPresignExpiration     = 7 * 24 * time.Hour
)

// s3API is the subset of *awss3.Client used by storage — enables mocking in tests.
type s3API interface {
	GetObject(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.Options)) (*awss3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error)
	HeadObject(ctx context.Context, params *awss3.HeadObjectInput, optFns ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error)
}

// presignAPI is the subset of *awss3.PresignClient used by storage.
type presignAPI interface {
	PresignPutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignGetObject(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type storage struct {
	client    s3API
	presigner presignAPI
	logger    logger.Logger
	tracer    trace.Tracer
}

// GetObject downloads an object from S3.
func (s *storage) GetObject(ctx context.Context, req GetObjectRequest) (*GetObjectResponse, error) {
	ctx, span := s.tracer.Start(ctx, "s3.get_object",
		trace.WithAttributes(
			attribute.String("s3.bucket", req.Bucket),
			attribute.String("s3.key", req.Key),
		))
	defer span.End()

	if err := validateBucketKey(req.Bucket, req.Key); err != nil {
		return nil, s.fail(span, err)
	}

	out, err := s.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(req.Bucket),
		Key:    aws.String(req.Key),
	})
	if err != nil {
		s.logError(ctx, "s3.get_object", "failed to get object", "bucket", req.Bucket, "key", req.Key, "error", err.Error())
		return nil, s.fail(span, fmt.Errorf("s3: get object: %w", err))
	}

	s.logInfo(ctx, "s3.get_object", "object retrieved", "bucket", req.Bucket, "key", req.Key)
	span.SetStatus(codes.Ok, "")

	return &GetObjectResponse{
		Body:          out.Body,
		ContentType:   derefString(out.ContentType),
		ContentLength: derefInt64(out.ContentLength),
		ETag:          derefString(out.ETag),
		LastModified:  derefTime(out.LastModified),
		Metadata:      out.Metadata,
	}, nil
}

// PutObject uploads an object to S3 directly from the backend.
func (s *storage) PutObject(ctx context.Context, req PutObjectRequest) (*PutObjectResult, error) {
	ctx, span := s.tracer.Start(ctx, "s3.put_object",
		trace.WithAttributes(
			attribute.String("s3.bucket", req.Bucket),
			attribute.String("s3.key", req.Key),
			attribute.String("s3.content_type", req.ContentType),
		))
	defer span.End()

	if err := validateBucketKey(req.Bucket, req.Key); err != nil {
		return nil, s.fail(span, err)
	}
	if req.Body == nil {
		return nil, s.fail(span, fmt.Errorf("s3: body is required"))
	}

	input := &awss3.PutObjectInput{
		Bucket: aws.String(req.Bucket),
		Key:    aws.String(req.Key),
		Body:   req.Body,
	}
	if req.ContentType != "" {
		input.ContentType = aws.String(req.ContentType)
	}
	if len(req.Metadata) > 0 {
		input.Metadata = req.Metadata
	}

	out, err := s.client.PutObject(ctx, input)
	if err != nil {
		s.logError(ctx, "s3.put_object", "failed to put object", "bucket", req.Bucket, "key", req.Key, "error", err.Error())
		return nil, s.fail(span, fmt.Errorf("s3: put object: %w", err))
	}

	result := &PutObjectResult{VersionId: derefString(out.VersionId)}
	s.logInfo(ctx, "s3.put_object", "object uploaded", "bucket", req.Bucket, "key", req.Key, "version_id", result.VersionId)
	span.SetStatus(codes.Ok, "")
	return result, nil
}

// DeleteObject removes an object. Missing keys are silently ignored.
func (s *storage) DeleteObject(ctx context.Context, bucket, key string) error {
	ctx, span := s.tracer.Start(ctx, "s3.delete_object",
		trace.WithAttributes(
			attribute.String("s3.bucket", bucket),
			attribute.String("s3.key", key),
		))
	defer span.End()

	if err := validateBucketKey(bucket, key); err != nil {
		return s.fail(span, err)
	}

	if _, err := s.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}); err != nil {
		s.logError(ctx, "s3.delete_object", "failed to delete object", "bucket", bucket, "key", key, "error", err.Error())
		return s.fail(span, fmt.Errorf("s3: delete object: %w", err))
	}

	s.logInfo(ctx, "s3.delete_object", "object deleted", "bucket", bucket, "key", key)
	span.SetStatus(codes.Ok, "")
	return nil
}

// HeadObject returns object metadata without downloading the body.
// Returns ErrObjectNotFound if the object does not exist.
func (s *storage) HeadObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	ctx, span := s.tracer.Start(ctx, "s3.head_object",
		trace.WithAttributes(
			attribute.String("s3.bucket", bucket),
			attribute.String("s3.key", key),
		))
	defer span.End()

	if err := validateBucketKey(bucket, key); err != nil {
		return nil, s.fail(span, err)
	}

	out, err := s.client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		var apiErr smithy.APIError
		if errors.As(err, &notFound) || (errors.As(err, &apiErr) && apiErr.ErrorCode() == "404") {
			return nil, s.fail(span, ErrObjectNotFound)
		}
		s.logError(ctx, "s3.head_object", "failed to head object", "bucket", bucket, "key", key, "error", err.Error())
		return nil, s.fail(span, fmt.Errorf("s3: head object: %w", err))
	}

	span.SetStatus(codes.Ok, "")
	return &ObjectInfo{
		ContentType:   derefString(out.ContentType),
		ContentLength: derefInt64(out.ContentLength),
		ETag:          derefString(out.ETag),
		LastModified:  derefTime(out.LastModified),
		Metadata:      out.Metadata,
	}, nil
}

// PresignPutObject generates a short-lived URL for client-side uploads.
func (s *storage) PresignPutObject(ctx context.Context, req PresignPutRequest) (*PresignedURL, error) {
	ctx, span := s.tracer.Start(ctx, "s3.presign_put",
		trace.WithAttributes(
			attribute.String("s3.bucket", req.Bucket),
			attribute.String("s3.key", req.Key),
		))
	defer span.End()

	if err := validateBucketKey(req.Bucket, req.Key); err != nil {
		return nil, s.fail(span, err)
	}

	expiresIn := clampExpiry(req.ExpiresIn)

	input := &awss3.PutObjectInput{
		Bucket: aws.String(req.Bucket),
		Key:    aws.String(req.Key),
	}
	if req.ContentType != "" {
		input.ContentType = aws.String(req.ContentType)
	}
	if len(req.Metadata) > 0 {
		input.Metadata = req.Metadata
	}

	presigned, err := s.presigner.PresignPutObject(ctx, input, awss3.WithPresignExpires(expiresIn))
	if err != nil {
		s.logError(ctx, "s3.presign_put", "failed to generate presigned PUT URL", "bucket", req.Bucket, "key", req.Key, "error", err.Error())
		return nil, s.fail(span, fmt.Errorf("s3: presign put: %w", err))
	}

	expiresAt := time.Now().Add(expiresIn)
	s.logInfo(ctx, "s3.presign_put", "presigned PUT URL generated", "bucket", req.Bucket, "key", req.Key, "expiresAt", expiresAt.Format(time.RFC3339))
	span.SetStatus(codes.Ok, "")

	return &PresignedURL{URL: presigned.URL, ExpiresAt: expiresAt}, nil
}

// PresignGetObject generates a short-lived URL for client-side downloads.
func (s *storage) PresignGetObject(ctx context.Context, req PresignGetRequest) (*PresignedURL, error) {
	ctx, span := s.tracer.Start(ctx, "s3.presign_get",
		trace.WithAttributes(
			attribute.String("s3.bucket", req.Bucket),
			attribute.String("s3.key", req.Key),
		))
	defer span.End()

	if err := validateBucketKey(req.Bucket, req.Key); err != nil {
		return nil, s.fail(span, err)
	}

	expiresIn := clampExpiry(req.ExpiresIn)

	presigned, err := s.presigner.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(req.Bucket),
		Key:    aws.String(req.Key),
	}, awss3.WithPresignExpires(expiresIn))
	if err != nil {
		s.logError(ctx, "s3.presign_get", "failed to generate presigned GET URL", "bucket", req.Bucket, "key", req.Key, "error", err.Error())
		return nil, s.fail(span, fmt.Errorf("s3: presign get: %w", err))
	}

	expiresAt := time.Now().Add(expiresIn)
	s.logInfo(ctx, "s3.presign_get", "presigned GET URL generated", "bucket", req.Bucket, "key", req.Key, "expiresAt", expiresAt.Format(time.RFC3339))
	span.SetStatus(codes.Ok, "")

	return &PresignedURL{URL: presigned.URL, ExpiresAt: expiresAt}, nil
}

// --- helpers ---

func (s *storage) logInfo(ctx context.Context, step, msg string, kv ...any) {
	if s.logger != nil {
		s.logger.Info(ctx, step, msg, kv...)
	}
}

func (s *storage) logError(ctx context.Context, step, msg string, kv ...any) {
	if s.logger != nil {
		s.logger.Error(ctx, step, msg, kv...)
	}
}

func (s *storage) fail(span trace.Span, err error) error {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}

func validateBucketKey(bucket, key string) error {
	if bucket == "" {
		return fmt.Errorf("s3: bucket is required")
	}
	if key == "" {
		return fmt.Errorf("s3: key is required")
	}
	return nil
}

func clampExpiry(d time.Duration) time.Duration {
	if d <= 0 {
		return defaultPresignExpiration
	}
	if d > maxPresignExpiration {
		return maxPresignExpiration
	}
	return d
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
