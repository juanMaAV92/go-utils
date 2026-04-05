# messaging/sns

SNS producer with OTel tracing and W3C Trace Context propagation.

> SNS is publish-only — there is no consumer. Subscribers receive messages via push (HTTP, Lambda, SQS, email). To consume SNS messages delivered through SQS, use `messaging/sqs/consumer` which automatically unwraps SNS envelopes.

## Setup

```go
import (
    "github.com/juanmaAV/go-utils/messaging/sns"
    "github.com/juanmaAV/go-utils/messaging/sns/producer"
)

snsCfg, err := sns.ConfigFromEnv("SNS")
client, err := sns.NewClient(ctx, snsCfg, logger)

prodCfg, err := producer.ConfigFromEnv("SNS")
prod, err := producer.New(client, logger, prodCfg, "order-producer")
```

### Config fields

**Base (`sns.ConfigFromEnv`)**

| Field | Env var (prefix=SNS) | Default |
|---|---|---|
| `Region` | `SNS_REGION` | required |
| `Endpoint` | `SNS_ENDPOINT` | empty (real AWS) |

**Producer (`producer.ConfigFromEnv`)**

| Field | Env var (prefix=SNS) | Default |
|---|---|---|
| `TopicArn` | `SNS_TOPIC_ARN` | required |

Multiple topics share one client — only the `TopicArn` differs:

```go
snsCfg, _ := sns.ConfigFromEnv("SNS")
client, _ := sns.NewClient(ctx, snsCfg, logger)

orderCfg,   _ := producer.ConfigFromEnv("ORDER_SNS")    // ORDER_SNS_TOPIC_ARN
invoiceCfg, _ := producer.ConfigFromEnv("INVOICE_SNS")  // INVOICE_SNS_TOPIC_ARN

orderProd,   _ := producer.New(client, logger, orderCfg, "order-producer")
invoiceProd, _ := producer.New(client, logger, invoiceCfg, "invoice-producer")
```

## Publish

```go
err := prod.Publish(ctx, &producer.Message{
    Body:       `{"order_id":"123","status":"placed"}`,
    Attributes: map[string]string{"source": "order-service"},
})
```

### Subject

Used by some subscription protocols (e.g. email):

```go
err := prod.Publish(ctx, &producer.Message{
    Body:    `{"order_id":"123"}`,
    Subject: "Order Placed",
})
```

### FIFO topics

```go
err := prod.Publish(ctx, &producer.Message{
    Body:                   `{"order_id":"123"}`,
    MessageGroupId:         "order-group",
    MessageDeduplicationId: "order-123-placed",
})
```

## Observability

- Every `Publish` creates an OTel span with `SpanKindProducer`
- W3C Trace Context (`traceparent`, `tracestate`) is injected into message attributes — distributed traces flow to downstream subscribers automatically
- Span attributes follow OpenTelemetry semantic conventions (`messaging.system=sns`, `messaging.operation=publish`, `messaging.destination.kind=topic`, etc.)

> **Attribute limit:** AWS SNS allows max 10 message attributes. 2 are reserved for W3C Trace Context. Keep user attributes ≤ 8.

## Notes

- For LocalStack: set `SNS_ENDPOINT=http://localhost:4566`
- Credentials are loaded from the standard AWS credential chain
