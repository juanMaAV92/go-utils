package s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.opentelemetry.io/otel"
)

// --- mocks ---

type mockS3 struct {
	getObjectFn    func(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.Options)) (*awss3.GetObjectOutput, error)
	putObjectFn    func(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error)
	deleteObjectFn func(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error)
	headObjectFn   func(ctx context.Context, params *awss3.HeadObjectInput, optFns ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error)
}

func (m *mockS3) GetObject(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
	return m.getObjectFn(ctx, params, optFns...)
}
func (m *mockS3) PutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
	return m.putObjectFn(ctx, params, optFns...)
}
func (m *mockS3) DeleteObject(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error) {
	return m.deleteObjectFn(ctx, params, optFns...)
}
func (m *mockS3) HeadObject(ctx context.Context, params *awss3.HeadObjectInput, optFns ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error) {
	return m.headObjectFn(ctx, params, optFns...)
}

type mockPresigner struct {
	presignPutFn func(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	presignGetFn func(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

func (m *mockPresigner) PresignPutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	return m.presignPutFn(ctx, params, optFns...)
}
func (m *mockPresigner) PresignGetObject(ctx context.Context, params *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	return m.presignGetFn(ctx, params, optFns...)
}

func strPtr(s string) *string { return &s }
func i64Ptr(i int64) *int64   { return &i }
func timePtr(t time.Time) *time.Time { return &t }

func newTestStorage(client s3API, presigner presignAPI) Storage {
	return &storage{
		client:    client,
		presigner: presigner,
		logger:    nil,
		tracer:    otel.Tracer("test"),
	}
}

var ctx = context.Background()

// ---- GetObject ----

func TestGetObject_Success(t *testing.T) {
	body := io.NopCloser(strings.NewReader("hello"))
	s := newTestStorage(&mockS3{
		getObjectFn: func(_ context.Context, params *awss3.GetObjectInput, _ ...func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
			return &awss3.GetObjectOutput{
				Body:          body,
				ContentType:   strPtr("text/plain"),
				ContentLength: i64Ptr(5),
				ETag:          strPtr(`"abc"`),
			}, nil
		},
	}, nil)

	resp, err := s.GetObject(ctx, GetObjectRequest{Bucket: "my-bucket", Key: "file.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want text/plain", resp.ContentType)
	}
	if resp.ContentLength != 5 {
		t.Errorf("ContentLength = %d, want 5", resp.ContentLength)
	}
}

func TestGetObject_MissingBucket(t *testing.T) {
	s := newTestStorage(&mockS3{}, nil)
	_, err := s.GetObject(ctx, GetObjectRequest{Bucket: "", Key: "k"})
	if err == nil {
		t.Error("expected error for missing bucket")
	}
}

func TestGetObject_MissingKey(t *testing.T) {
	s := newTestStorage(&mockS3{}, nil)
	_, err := s.GetObject(ctx, GetObjectRequest{Bucket: "b", Key: ""})
	if err == nil {
		t.Error("expected error for missing key")
	}
}

// ---- PutObject ----

func TestPutObject_Success(t *testing.T) {
	var capturedKey string
	s := newTestStorage(&mockS3{
		putObjectFn: func(_ context.Context, params *awss3.PutObjectInput, _ ...func(*awss3.Options)) (*awss3.PutObjectOutput, error) {
			capturedKey = *params.Key
			return &awss3.PutObjectOutput{}, nil
		},
	}, nil)

	err := s.PutObject(ctx, PutObjectRequest{
		Bucket:      "my-bucket",
		Key:         "uploads/file.txt",
		Body:        bytes.NewReader([]byte("data")),
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedKey != "uploads/file.txt" {
		t.Errorf("key = %q, want uploads/file.txt", capturedKey)
	}
}

func TestPutObject_NilBody(t *testing.T) {
	s := newTestStorage(&mockS3{}, nil)
	err := s.PutObject(ctx, PutObjectRequest{Bucket: "b", Key: "k", Body: nil})
	if err == nil {
		t.Error("expected error for nil body")
	}
}

// ---- DeleteObject ----

func TestDeleteObject_Success(t *testing.T) {
	var deletedKey string
	s := newTestStorage(&mockS3{
		deleteObjectFn: func(_ context.Context, params *awss3.DeleteObjectInput, _ ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error) {
			deletedKey = *params.Key
			return &awss3.DeleteObjectOutput{}, nil
		},
	}, nil)

	if err := s.DeleteObject(ctx, "my-bucket", "file.txt"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedKey != "file.txt" {
		t.Errorf("key = %q, want file.txt", deletedKey)
	}
}

// ---- HeadObject ----

func TestHeadObject_Success(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	s := newTestStorage(&mockS3{
		headObjectFn: func(_ context.Context, params *awss3.HeadObjectInput, _ ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error) {
			return &awss3.HeadObjectOutput{
				ContentType:   strPtr("image/png"),
				ContentLength: i64Ptr(1024),
				ETag:          strPtr(`"etag123"`),
				LastModified:  timePtr(now),
			}, nil
		},
	}, nil)

	info, err := s.HeadObject(ctx, "my-bucket", "image.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ContentType != "image/png" {
		t.Errorf("ContentType = %q, want image/png", info.ContentType)
	}
	if info.ContentLength != 1024 {
		t.Errorf("ContentLength = %d, want 1024", info.ContentLength)
	}
}

func TestHeadObject_NotFound(t *testing.T) {
	s := newTestStorage(&mockS3{
		headObjectFn: func(_ context.Context, _ *awss3.HeadObjectInput, _ ...func(*awss3.Options)) (*awss3.HeadObjectOutput, error) {
			return nil, &types.NotFound{}
		},
	}, nil)

	_, err := s.HeadObject(ctx, "my-bucket", "missing.txt")
	if !errors.Is(err, ErrObjectNotFound) {
		t.Errorf("expected ErrObjectNotFound, got %v", err)
	}
}

// ---- Presign ----

func TestPresignPutObject_DefaultExpiry(t *testing.T) {
	var capturedExpiry time.Duration
	s := newTestStorage(nil, &mockPresigner{
		presignPutFn: func(_ context.Context, _ *awss3.PutObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
			o := &awss3.PresignOptions{}
			for _, fn := range optFns {
				fn(o)
			}
			capturedExpiry = o.Expires
			return &v4.PresignedHTTPRequest{URL: "https://example.com/upload"}, nil
		},
	})

	result, err := s.PresignPutObject(ctx, PresignPutRequest{Bucket: "b", Key: "k"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "https://example.com/upload" {
		t.Errorf("URL = %q", result.URL)
	}
	if capturedExpiry != defaultPresignExpiration {
		t.Errorf("expiry = %v, want %v", capturedExpiry, defaultPresignExpiration)
	}
}

func TestPresignGetObject_CustomExpiry(t *testing.T) {
	var capturedExpiry time.Duration
	s := newTestStorage(nil, &mockPresigner{
		presignGetFn: func(_ context.Context, _ *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
			o := &awss3.PresignOptions{}
			for _, fn := range optFns {
				fn(o)
			}
			capturedExpiry = o.Expires
			return &v4.PresignedHTTPRequest{URL: "https://example.com/download"}, nil
		},
	})

	result, err := s.PresignGetObject(ctx, PresignGetRequest{Bucket: "b", Key: "k", ExpiresIn: time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedExpiry != time.Hour {
		t.Errorf("expiry = %v, want 1h", capturedExpiry)
	}
	if result.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}
}

func TestPresignGetObject_CapsAtMaxExpiry(t *testing.T) {
	var capturedExpiry time.Duration
	s := newTestStorage(nil, &mockPresigner{
		presignGetFn: func(_ context.Context, _ *awss3.GetObjectInput, optFns ...func(*awss3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
			o := &awss3.PresignOptions{}
			for _, fn := range optFns {
				fn(o)
			}
			capturedExpiry = o.Expires
			return &v4.PresignedHTTPRequest{URL: "https://example.com/download"}, nil
		},
	})

	_, err := s.PresignGetObject(ctx, PresignGetRequest{Bucket: "b", Key: "k", ExpiresIn: 30 * 24 * time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedExpiry != maxPresignExpiration {
		t.Errorf("expiry = %v, want %v (max)", capturedExpiry, maxPresignExpiration)
	}
}

// ---- clampExpiry ----

func TestClampExpiry(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want time.Duration
	}{
		{0, defaultPresignExpiration},
		{-time.Minute, defaultPresignExpiration},
		{time.Hour, time.Hour},
		{30 * 24 * time.Hour, maxPresignExpiration},
	}
	for _, tc := range cases {
		got := clampExpiry(tc.in)
		if got != tc.want {
			t.Errorf("clampExpiry(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
