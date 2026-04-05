package scheduler

import (
	"context"
	"time"
)

// Scheduler is the interface for EventBridge Scheduler operations.
type Scheduler interface {
	// CreateSchedule creates a one-time schedule.
	CreateSchedule(ctx context.Context, cfg ScheduleConfig) (*ScheduleResult, error)

	// UpdateSchedule updates an existing schedule in-place.
	UpdateSchedule(ctx context.Context, cfg ScheduleConfig) (*ScheduleResult, error)

	// DeleteSchedule removes a schedule. groupName defaults to "default" if empty.
	DeleteSchedule(ctx context.Context, name, groupName string) error

	// GetSchedule retrieves schedule metadata. groupName defaults to "default" if empty.
	GetSchedule(ctx context.Context, name, groupName string) (*ScheduleInfo, error)
}

// ScheduleConfig holds the parameters for creating or updating a schedule.
type ScheduleConfig struct {
	Name                      string
	Description               string
	GroupName                 string // default: "default"
	ScheduleTime              time.Time
	Timezone                  string // IANA timezone, default: "UTC"
	FlexibleTimeWindowMinutes int    // 0 = exact time, >0 = flexible window in minutes
	Target                    ScheduleTarget
	Tags                      map[string]string
}

// NewWebhookSchedule is a convenience constructor for scheduling an HTTP webhook
// via a Lambda proxy function.
func NewWebhookSchedule(
	name, description, groupName string,
	scheduledTime time.Time,
	lambdaProxyArn, webhookURL, method string,
	payload any,
) ScheduleConfig {
	return ScheduleConfig{
		Name:         name,
		Description:  description,
		GroupName:    groupName,
		ScheduleTime: scheduledTime,
		Timezone:     "UTC",
		Target: &LambdaTarget{
			LambdaArn: lambdaProxyArn,
			Payload: map[string]any{
				"url":    webhookURL,
				"method": method,
				"body":   payload,
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
	}
}

// ScheduleTarget is implemented by all supported target types.
type ScheduleTarget interface {
	isScheduleTarget()
}

// LambdaTarget invokes a Lambda function.
type LambdaTarget struct {
	LambdaArn   string
	Payload     map[string]any
	RetryPolicy *RetryPolicy // nil = defaults (3 attempts, 1h max age)
}

func (LambdaTarget) isScheduleTarget() {}

// RetryPolicy controls retry behavior for a schedule target.
type RetryPolicy struct {
	MaxAttempts     int32 // default: 3
	MaxEventAgeSecs int32 // default: 3600 (1 hour)
}

// ScheduleResult is returned by CreateSchedule and UpdateSchedule.
type ScheduleResult struct {
	ScheduleName string
	ScheduleArn  string
}

// ScheduleInfo is returned by GetSchedule.
type ScheduleInfo struct {
	Name      string
	Arn       string
	State     string // "ENABLED" or "DISABLED"
	GroupName string
}
