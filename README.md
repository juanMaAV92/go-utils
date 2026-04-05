# go-utils

Go utility library for building microservices on AWS. Single module, consistent patterns across all packages: `ConfigFromEnv`, interface-driven design, OTel tracing.

```bash
go get github.com/juanmaAV/go-utils
```

Requires Go 1.21+.

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
