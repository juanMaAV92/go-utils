# messaging/scheduler

AWS EventBridge Scheduler client with OTel tracing.

> Schedules one-time invocations at a specific UTC time. Currently supports Lambda as the only target type.

## Setup

```go
import "github.com/juanmaAV/go-utils/messaging/scheduler"

cfg, err := scheduler.ConfigFromEnv("SCHEDULER")
sched, err := scheduler.New(ctx, cfg, logger)
```

### Config fields

| Field | Env var (prefix=SCHEDULER) | Default |
|---|---|---|
| `Region` | `SCHEDULER_REGION` | required |
| `RoleArn` | `SCHEDULER_ROLE_ARN` | required |
| `Endpoint` | `SCHEDULER_ENDPOINT` | empty (real AWS) |

`RoleArn` is the IAM role EventBridge Scheduler assumes when invoking targets.

## Create a schedule

```go
result, err := sched.CreateSchedule(ctx, scheduler.ScheduleConfig{
    Name:         "send-reminder-abc123",
    Description:  "Send reminder for order abc123",
    GroupName:    "reminders",              // default: "default"
    ScheduleTime: time.Now().Add(2 * time.Hour),
    Target: &scheduler.LambdaTarget{
        LambdaArn: "arn:aws:lambda:us-east-1:123456789012:function:my-fn",
        Payload:   map[string]any{"order_id": "abc123"},
    },
})
```

### Flexible time window

```go
scheduler.ScheduleConfig{
    FlexibleTimeWindowMinutes: 15, // invoke within a 15-minute window
    ...
}
```

### Retry policy

```go
&scheduler.LambdaTarget{
    LambdaArn: "...",
    Payload:   map[string]any{},
    RetryPolicy: &scheduler.RetryPolicy{
        MaxAttempts:     5,
        MaxEventAgeSecs: 600,
    },
}
// defaults: MaxAttempts=3, MaxEventAgeSecs=3600
```

## Update a schedule

Same signature as create — replaces the existing schedule in-place:

```go
result, err := sched.UpdateSchedule(ctx, scheduler.ScheduleConfig{...})
```

## Delete a schedule

```go
err := sched.DeleteSchedule(ctx, "send-reminder-abc123", "reminders")
// groupName defaults to "default" if empty
```

## Get a schedule

```go
info, err := sched.GetSchedule(ctx, "send-reminder-abc123", "reminders")
// info.State → "ENABLED" or "DISABLED"
```

## HTTP webhook via Lambda proxy

`NewWebhookSchedule` is a convenience constructor for scheduling an HTTP call through a Lambda proxy function:

```go
cfg := scheduler.NewWebhookSchedule(
    "notify-order-abc123",
    "Notify order service",
    "notifications",
    time.Now().Add(1*time.Hour),
    "arn:aws:lambda:us-east-1:123456789012:function:http-proxy",
    "https://api.example.com/orders/abc123/notify",
    "POST",
    map[string]any{"status": "shipped"},
)
result, err := sched.CreateSchedule(ctx, cfg)
```

The Lambda proxy receives:
```json
{
  "url": "https://api.example.com/orders/abc123/notify",
  "method": "POST",
  "body": {"status": "shipped"},
  "headers": {"Content-Type": "application/json"}
}
```

## Observability

- Every operation creates an OTel span (`scheduler.create`, `scheduler.update`, `scheduler.delete`, `scheduler.get`)
- Span attributes: `schedule.name`, `schedule.time`
- Errors are recorded on the span with status `Error`

## Notes

- `ScheduleTime` must be in the future — `CreateSchedule` and `UpdateSchedule` return an error if it's in the past
- For LocalStack: set `SCHEDULER_ENDPOINT=http://localhost:4566`
- Credentials are loaded from the standard AWS credential chain
