# messaging/sqs

SQS producer and consumer with OTel tracing, W3C Trace Context propagation, SNS envelope unwrapping, and worker pool.

## Setup

```go
import (
    "github.com/juanmaAV/go-utils/messaging/sqs"
    "github.com/juanmaAV/go-utils/messaging/sqs/producer"
    "github.com/juanmaAV/go-utils/messaging/sqs/consumer"
)

// 1. Create the base SQS client (shared by producer and consumer)
sqsCfg, err := sqs.ConfigFromEnv("SQS")
client, err := sqs.NewClient(ctx, sqsCfg, logger)

// 2. Producer
prodCfg, err := producer.ConfigFromEnv("SQS")
prod, err := producer.New(client, logger, prodCfg, "order-producer")

// 3. Consumer
consCfg, err := consumer.ConfigFromEnv("SQS")
cons, err := consumer.New(client, myProcessor, logger, consCfg, "order-consumer")
```

### Config fields

**Base (`sqs.ConfigFromEnv`)**

| Field | Env var (prefix=SQS) | Default |
|---|---|---|
| `Region` | `SQS_REGION` | required |
| `Endpoint` | `SQS_ENDPOINT` | empty (real AWS) |

**Producer (`producer.ConfigFromEnv`)**

| Field | Env var (prefix=SQS) | Default |
|---|---|---|
| `QueueURL` | `SQS_QUEUE_URL` | required |

**Consumer (`consumer.ConfigFromEnv`)**

| Field | Env var (prefix=SQS) | Default |
|---|---|---|
| `QueueURL` | `SQS_QUEUE_URL` | required |
| `MaxMessages` | `SQS_MAX_MESSAGES` | `10` |
| `WaitTimeSeconds` | `SQS_WAIT_TIME_SECONDS` | `20` |
| `VisibilityTimeout` | `SQS_VISIBILITY_TIMEOUT` | `30` |
| `WorkerPoolSize` | `SQS_WORKER_POOL_SIZE` | `10` |

**Multiple queues in the same region share one client** — only the queue URL differs per producer/consumer:

```go
// One client, shared across all queues in the same region
sqsCfg, _ := sqs.ConfigFromEnv("SQS")
client, _ := sqs.NewClient(ctx, sqsCfg, logger)

// Each producer/consumer gets its own queue URL via a different prefix
orderCfg,   _ := producer.ConfigFromEnv("ORDER_SQS")    // ORDER_SQS_QUEUE_URL
invoiceCfg, _ := producer.ConfigFromEnv("INVOICE_SQS")  // INVOICE_SQS_QUEUE_URL

orderProd,   _ := producer.New(client, logger, orderCfg, "order-producer")
invoiceProd, _ := producer.New(client, logger, invoiceCfg, "invoice-producer")
```

A separate `sqs.ConfigFromEnv` prefix is only needed when queues are in **different regions or endpoints** (e.g., multi-region setup or LocalStack vs real AWS).

## Producer

### SendMessage

```go
err := prod.SendMessage(ctx, &producer.Message{
    Body:       `{"order_id":"123","status":"placed"}`,
    Attributes: map[string]string{"source": "order-service"},
})
```

**FIFO queues** — set `MessageGroupId` and optionally `MessageDeduplicationId`:

```go
err := prod.SendMessage(ctx, &producer.Message{
    Body:                   `{"order_id":"123"}`,
    MessageGroupId:         "order-group",
    MessageDeduplicationId: "order-123-placed",
})
```

### SendBatch

Sends up to 10 messages in a single request. Returns `*BatchResult` even on partial failure.

```go
msgs := []*producer.Message{
    {Body: `{"id":"1"}`},
    {Body: `{"id":"2"}`},
    {Body: `{"id":"3"}`},
}

result, err := prod.SendBatch(ctx, msgs)
// result.SuccessCount, result.FailedCount, result.FailedIds
if err != nil {
    // partial failure — some messages were sent, check result for details
}
```

> **Attribute limit:** AWS SQS allows max 10 message attributes. 2 are reserved for W3C Trace Context (`traceparent`, `tracestate`). Keep user attributes ≤ 8.

## Consumer

Implement `MessageProcessor` and call `Start`:

```go
type OrderProcessor struct{}

func (p *OrderProcessor) ProcessMessage(ctx context.Context, body []byte) error {
    var order Order
    if err := json.Unmarshal(body, &order); err != nil {
        return err // message will be retried
    }
    return processOrder(ctx, order) // return nil to delete from queue
}

// Start is blocking — run in a goroutine or as the main loop
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

if err := cons.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
    log.Fatal(err)
}
```

### Return semantics

| Processor return | SQS behavior |
|---|---|
| `nil` | Message deleted from queue |
| `error` | Message left in queue — redelivered after visibility timeout |

### SNS → SQS fan-out

The consumer automatically detects and unwraps SNS envelopes. No configuration needed — the same consumer handles both direct SQS messages and SNS-proxied messages transparently.

```
SNS topic → SQS subscription → consumer detects SNS wrapper → passes inner message to ProcessMessage
```

## Observability

- Every `SendMessage`, `SendBatch`, and `ProcessMessage` creates an OTel span with `SpanKindProducer` / `SpanKindConsumer`
- W3C Trace Context (`traceparent`, `tracestate`) is injected into message attributes on send and extracted on receive — distributed traces flow across services automatically
- Span attributes follow OpenTelemetry semantic conventions for messaging (`messaging.system=sqs`, `messaging.operation`, `messaging.destination.name`, etc.)

## Notes

- `Start` blocks until ctx is cancelled — run in a goroutine or as the main service loop
- For LocalStack: set `SQS_ENDPOINT=http://localhost:4566`
