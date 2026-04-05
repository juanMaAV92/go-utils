package s3

import (
	"context"
	"io"
	"time"
)

// Storage is the interface for S3 operations.
type Storage interface {
	// GetObject downloads an object and returns its body and metadata.
	GetObject(ctx context.Context, req GetObjectRequest) (*GetObjectResponse, error)

	// PutObject uploads an object directly from the backend.
	PutObject(ctx context.Context, req PutObjectRequest) error

	// DeleteObject removes an object. Missing keys are silently ignored.
	DeleteObject(ctx context.Context, bucket, key string) error

	// HeadObject returns metadata without downloading the body.
	// Returns ErrObjectNotFound if the object does not exist.
	HeadObject(ctx context.Context, bucket, key string) (*ObjectInfo, error)

	// PresignPutObject generates a short-lived URL for client-side uploads.
	PresignPutObject(ctx context.Context, req PresignPutRequest) (*PresignedURL, error)

	// PresignGetObject generates a short-lived URL for client-side downloads.
	PresignGetObject(ctx context.Context, req PresignGetRequest) (*PresignedURL, error)
}

// GetObjectRequest contains parameters for downloading an object.
type GetObjectRequest struct {
	Bucket string
	Key    string
}

// GetObjectResponse contains the downloaded object body and metadata.
type GetObjectResponse struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	ETag          string
	LastModified  time.Time
	Metadata      map[string]string
}

// PutObjectRequest contains parameters for uploading an object.
type PutObjectRequest struct {
	Bucket      string
	Key         string
	Body        io.Reader
	ContentType string            // optional
	Metadata    map[string]string // optional
}

// ObjectInfo contains metadata returned by HeadObject.
type ObjectInfo struct {
	ContentType   string
	ContentLength int64
	ETag          string
	LastModified  time.Time
	Metadata      map[string]string
}

// PresignPutRequest contains parameters for generating a presigned PUT URL.
type PresignPutRequest struct {
	Bucket      string
	Key         string
	ContentType string            // optional
	Metadata    map[string]string // optional
	ExpiresIn   time.Duration     // default: 5 minutes, max: 7 days
}

// PresignGetRequest contains parameters for generating a presigned GET URL.
type PresignGetRequest struct {
	Bucket    string
	Key       string
	ExpiresIn time.Duration // default: 5 minutes, max: 7 days
}

// PresignedURL is the result of a presign operation.
type PresignedURL struct {
	URL       string
	ExpiresAt time.Time
}
