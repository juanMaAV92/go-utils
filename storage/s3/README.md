# storage/s3

S3 client with OTel tracing and structured logging. Supports direct backend operations (Get, Put, Delete, Head) and presigned URLs for client-side uploads/downloads.

## Setup

```go
import "github.com/juanmaAV/go-utils/storage/s3"

cfg, err := s3.ConfigFromEnv("S3")
if err != nil {
    log.Fatal(err)
}

store, err := s3.New(ctx, cfg, logger)
if err != nil {
    log.Fatal(err)
}
```

### Config fields

| Field | Env var (prefix=S3) | Default |
|---|---|---|
| `Region` | `S3_REGION` | required |
| `Endpoint` | `S3_ENDPOINT` | empty (real AWS) |

`Endpoint` is only needed for LocalStack or custom S3-compatible endpoints.

Multiple S3 buckets sharing the same client is fine — bucket is passed per-call, not in config.

Multiple S3 configurations (e.g., different regions):

```go
mediaCfg, _    := s3.ConfigFromEnv("MEDIA_S3")
archiveCfg, _  := s3.ConfigFromEnv("ARCHIVE_S3")
```

## Sentinel errors

```go
_, err := store.HeadObject(ctx, "my-bucket", "missing.txt")
if errors.Is(err, s3.ErrObjectNotFound) {
    // object does not exist
}
```

## Operations

### GetObject

Downloads an object. The caller is responsible for closing `resp.Body`.

```go
resp, err := store.GetObject(ctx, s3.GetObjectRequest{
    Bucket: "my-bucket",
    Key:    "uploads/file.pdf",
})
if err != nil {
    return err
}
defer resp.Body.Close()

// resp.ContentType, resp.ContentLength, resp.ETag, resp.LastModified, resp.Metadata
```

### PutObject

Direct backend upload — use this when the server controls the file content.

```go
f, _ := os.Open("report.pdf")
defer f.Close()

err := store.PutObject(ctx, s3.PutObjectRequest{
    Bucket:      "my-bucket",
    Key:         "reports/2024/report.pdf",
    Body:        f,
    ContentType: "application/pdf",
    Metadata:    map[string]string{"uploaded-by": "service-a"},
})
```

### DeleteObject

Missing keys are silently ignored (S3 behavior).

```go
err := store.DeleteObject(ctx, "my-bucket", "uploads/old-file.txt")
```

### HeadObject

Returns metadata without downloading the body. Useful to check existence or get size before downloading.

```go
info, err := store.HeadObject(ctx, "my-bucket", "uploads/file.pdf")
if errors.Is(err, s3.ErrObjectNotFound) {
    // does not exist
}
// info.ContentType, info.ContentLength, info.ETag, info.LastModified, info.Metadata
```

### PresignPutObject

Generates a short-lived URL for client-side uploads (e.g., browser → S3 directly, bypassing your backend).

```go
url, err := store.PresignPutObject(ctx, s3.PresignPutRequest{
    Bucket:      "my-bucket",
    Key:         "uploads/user-123/avatar.jpg",
    ContentType: "image/jpeg",
    ExpiresIn:   15 * time.Minute,
})
// url.URL, url.ExpiresAt
```

Default expiry: **5 minutes**. Maximum: **7 days**.

### PresignGetObject

Generates a short-lived URL for client-side downloads.

```go
url, err := store.PresignGetObject(ctx, s3.PresignGetRequest{
    Bucket:    "my-bucket",
    Key:       "reports/2024/report.pdf",
    ExpiresIn: time.Hour,
})
```

## Notes

- OTel tracing is enabled automatically on every operation
- `DeleteObject` on a non-existent key returns no error (S3 behavior)
- For LocalStack: set `S3_ENDPOINT=http://localhost:4566`
- Credentials are loaded from the standard AWS credential chain (env vars, `~/.aws/credentials`, IAM role)
