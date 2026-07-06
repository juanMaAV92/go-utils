# go-utils

[![CI](https://github.com/juanMaAV92/go-utils/actions/workflows/ci.yml/badge.svg)](https://github.com/juanMaAV92/go-utils/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/juanMaAV92/go-utils.svg)](https://pkg.go.dev/github.com/juanMaAV92/go-utils)
[![Go Report Card](https://goreportcard.com/badge/github.com/juanMaAV92/go-utils)](https://goreportcard.com/report/github.com/juanMaAV92/go-utils)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Go utility library for building microservices on AWS. Single module, consistent patterns across all packages: `ConfigFromEnv`, interface-driven design, OTel tracing.

📖 **[Interactive docs & module guide](https://juanmaav92.github.io/go-utils)** · real-world usage: **[go-echo-blueprint](https://github.com/juanMaAV92/go-echo-blueprint)**

```bash
go get github.com/juanMaAV92/go-utils/v2
```

```go
import "github.com/juanMaAV92/go-utils/v2/logger"
```

Requires Go 1.25+.

---

## v2.0.0

Module path is now `/v2` (Go semantic import versioning). Update imports to `github.com/juanMaAV92/go-utils/v2/...`. Highlights of this release:

- **logger** — OTel `trace_id`/`span_id` are now actually injected into every log line (previously silently dropped by a handler-wrapping bug).
- **telemetry** — fractional sampling is `ParentBased`, so upstream sampling decisions are honored across services.
- **security/jwt** — validation now enforces RS256 only, requires `exp`, and validates the issuer.
- **httpclient** — request/response bodies are no longer logged (they leaked credentials and tokens).
- **cache/redis** — `WithKeepTTL` works (was a guaranteed syntax error); `AddToSet` rejects unsupported options.
- **messaging/scheduler** — `UpdateSchedule` no longer re-enables paused schedules; timezone is applied correctly; the non-functional `Tags` field was removed in favor of a `Disabled` flag.
- **messaging/sqs/consumer** — poll backoff on errors, graceful drain on shutdown (no duplicate delivery), `MaxMessages`/`WaitTimeSeconds` clamped to AWS limits.
- **database/postgresql** — errors wrap their cause (`errors.Is`-able); separate `MaxIdleConns`; higher default pool size.
- **middleware/identity** — scoped permission forwarding no longer leaks `all:all`.

---

## Packages

### Core

| Package | Description |
|---|---|
| [`env`](env/) | Environment variable parsing with type conversion and safe defaults |
| [`errors`](errors/) | Structured HTTP error responses; `errors/echo` for Echo error handler |
| [`logger`](logger/) | `log/slog`-based structured logger with OTel trace/span injection |
| [`telemetry`](telemetry/) | OpenTelemetry SDK initialisation (OTLP exporter, sampler, resource) |
| [`validator`](validator/) | `go-playground/validator` wrapper returning structured error responses |
| [`pointers`](pointers/) | Generic pointer helpers (`Pointer[T]`, `Value[T]`, `FirstNonNil`) |
| [`timeutil`](timeutil/) | Time formatting and pointer conversion utilities |
| [`httpclient`](httpclient/) | HTTP client (Resty) with OTel trace propagation and retry |
| [`security/jwt`](security/jwt/) | RS256 JWT generation and validation with generic claims |

### Infrastructure

| Package | Description |
|---|---|
| [`database/postgresql`](database/postgresql/) | GORM wrapper with CRUD, pagination, transactions, and OTel tracing |
| [`cache/redis`](cache/redis/) | Redis client with TTL, set operations, Pub/Sub, and OTel metrics |
| [`storage/s3`](storage/s3/) | S3 client: `GetObject`, `PutObject`, `DeleteObject`, `HeadObject`, presigned URLs |

### Messaging

| Package | Description |
|---|---|
| [`messaging/sqs`](messaging/sqs/) | SQS client; `sqs/producer` (send/batch), `sqs/consumer` (worker pool, SNS unwrap) |
| [`messaging/sns`](messaging/sns/) | SNS producer with W3C Trace Context propagation |
| [`messaging/scheduler`](messaging/scheduler/) | EventBridge Scheduler: one-time Lambda invocations, flexible windows, retry policy |

### Middleware & Testing

| Package | Description |
|---|---|
| [`middleware/identity`](middleware/identity/) | Echo middleware for user identity propagation via HTTP headers; RBAC helpers |
| [`testutil/echo`](testutil/echo/) | Table-driven Echo handler test helpers (`PrepareContext`, `ToJSONString`) |
| [`testutil/http`](testutil/http/) | Framework-agnostic HTTP test helpers (`AssertStatus`, `AssertJSONField`, `DecodeJSON`) |

---

## Design

- **Observability first** — every package emits OTel spans; logger injects `trace_id`/`span_id`
- **ConfigFromEnv(prefix)** — consistent across all AWS packages; prefix isolates env vars per client instance
- **Interface-driven** — exported interface, unexported implementation; internal API interfaces enable mock-based tests without real AWS
- **No hardcoded credentials** — all AWS packages use the standard credential chain

## License

MIT
