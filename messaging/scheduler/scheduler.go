package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssched "github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/juanMaAV92/go-utils/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultGroupName    = "default"
	defaultTimezone     = "UTC"
	defaultMaxAttempts  = int32(3)
	defaultMaxAgeSecs   = int32(3600)
)

// schedulerAPI is the subset of *awssched.Client used by sched — enables mocking in tests.
type schedulerAPI interface {
	CreateSchedule(ctx context.Context, params *awssched.CreateScheduleInput, optFns ...func(*awssched.Options)) (*awssched.CreateScheduleOutput, error)
	UpdateSchedule(ctx context.Context, params *awssched.UpdateScheduleInput, optFns ...func(*awssched.Options)) (*awssched.UpdateScheduleOutput, error)
	DeleteSchedule(ctx context.Context, params *awssched.DeleteScheduleInput, optFns ...func(*awssched.Options)) (*awssched.DeleteScheduleOutput, error)
	GetSchedule(ctx context.Context, params *awssched.GetScheduleInput, optFns ...func(*awssched.Options)) (*awssched.GetScheduleOutput, error)
}

type sched struct {
	client  schedulerAPI
	logger  logger.Logger
	tracer  trace.Tracer
	roleArn string
}

// CreateSchedule creates a one-time schedule.
func (s *sched) CreateSchedule(ctx context.Context, cfg ScheduleConfig) (*ScheduleResult, error) {
	ctx, span := s.tracer.Start(ctx, "scheduler.create",
		trace.WithAttributes(
			attribute.String("schedule.name", cfg.Name),
			attribute.String("schedule.time", cfg.ScheduleTime.Format(time.RFC3339)),
		))
	defer span.End()

	if err := validate(cfg); err != nil {
		return nil, s.fail(span, err)
	}

	input, err := s.buildCreateInput(cfg)
	if err != nil {
		return nil, s.fail(span, err)
	}

	out, err := s.client.CreateSchedule(ctx, input)
	if err != nil {
		s.logError(ctx, "scheduler.create", "failed to create schedule", "name", cfg.Name, "error", err.Error())
		return nil, s.fail(span, fmt.Errorf("scheduler: create: %w", err))
	}

	arn := aws.ToString(out.ScheduleArn)
	s.logInfo(ctx, "scheduler.create", "schedule created", "name", cfg.Name, "arn", arn)
	span.SetStatus(codes.Ok, "")
	return &ScheduleResult{ScheduleName: cfg.Name, ScheduleArn: arn}, nil
}

// UpdateSchedule updates an existing schedule in-place.
func (s *sched) UpdateSchedule(ctx context.Context, cfg ScheduleConfig) (*ScheduleResult, error) {
	ctx, span := s.tracer.Start(ctx, "scheduler.update",
		trace.WithAttributes(
			attribute.String("schedule.name", cfg.Name),
			attribute.String("schedule.time", cfg.ScheduleTime.Format(time.RFC3339)),
		))
	defer span.End()

	if err := validate(cfg); err != nil {
		return nil, s.fail(span, err)
	}

	input, err := s.buildUpdateInput(cfg)
	if err != nil {
		return nil, s.fail(span, err)
	}

	out, err := s.client.UpdateSchedule(ctx, input)
	if err != nil {
		s.logError(ctx, "scheduler.update", "failed to update schedule", "name", cfg.Name, "error", err.Error())
		return nil, s.fail(span, fmt.Errorf("scheduler: update: %w", err))
	}

	arn := aws.ToString(out.ScheduleArn)
	s.logInfo(ctx, "scheduler.update", "schedule updated", "name", cfg.Name, "arn", arn)
	span.SetStatus(codes.Ok, "")
	return &ScheduleResult{ScheduleName: cfg.Name, ScheduleArn: arn}, nil
}

// DeleteSchedule removes a schedule.
func (s *sched) DeleteSchedule(ctx context.Context, name, groupName string) error {
	ctx, span := s.tracer.Start(ctx, "scheduler.delete",
		trace.WithAttributes(attribute.String("schedule.name", name)))
	defer span.End()

	if name == "" {
		return s.fail(span, fmt.Errorf("scheduler: name is required"))
	}
	if groupName == "" {
		groupName = defaultGroupName
	}

	if _, err := s.client.DeleteSchedule(ctx, &awssched.DeleteScheduleInput{
		Name:      aws.String(name),
		GroupName: aws.String(groupName),
	}); err != nil {
		s.logError(ctx, "scheduler.delete", "failed to delete schedule", "name", name, "error", err.Error())
		return s.fail(span, fmt.Errorf("scheduler: delete: %w", err))
	}

	s.logInfo(ctx, "scheduler.delete", "schedule deleted", "name", name)
	span.SetStatus(codes.Ok, "")
	return nil
}

// GetSchedule retrieves schedule metadata.
func (s *sched) GetSchedule(ctx context.Context, name, groupName string) (*ScheduleInfo, error) {
	ctx, span := s.tracer.Start(ctx, "scheduler.get",
		trace.WithAttributes(attribute.String("schedule.name", name)))
	defer span.End()

	if name == "" {
		return nil, s.fail(span, fmt.Errorf("scheduler: name is required"))
	}
	if groupName == "" {
		groupName = defaultGroupName
	}

	out, err := s.client.GetSchedule(ctx, &awssched.GetScheduleInput{
		Name:      aws.String(name),
		GroupName: aws.String(groupName),
	})
	if err != nil {
		s.logError(ctx, "scheduler.get", "failed to get schedule", "name", name, "error", err.Error())
		return nil, s.fail(span, fmt.Errorf("scheduler: get: %w", err))
	}

	span.SetStatus(codes.Ok, "")
	return &ScheduleInfo{
		Name:      aws.ToString(out.Name),
		Arn:       aws.ToString(out.Arn),
		State:     string(out.State),
		GroupName: aws.ToString(out.GroupName),
	}, nil
}

// --- builders ---

func (s *sched) buildCreateInput(cfg ScheduleConfig) (*awssched.CreateScheduleInput, error) {
	ftw, expr, target, err := s.buildCommon(cfg)
	if err != nil {
		return nil, err
	}

	groupName := orDefault(cfg.GroupName, defaultGroupName)
	timezone := orDefault(cfg.Timezone, defaultTimezone)

	return &awssched.CreateScheduleInput{
		Name:                       aws.String(cfg.Name),
		Description:                aws.String(cfg.Description),
		GroupName:                  aws.String(groupName),
		ScheduleExpression:         aws.String(expr),
		ScheduleExpressionTimezone: aws.String(timezone),
		FlexibleTimeWindow:         ftw,
		Target:                     target,
	}, nil
}

func (s *sched) buildUpdateInput(cfg ScheduleConfig) (*awssched.UpdateScheduleInput, error) {
	ftw, expr, target, err := s.buildCommon(cfg)
	if err != nil {
		return nil, err
	}

	groupName := orDefault(cfg.GroupName, defaultGroupName)
	timezone := orDefault(cfg.Timezone, defaultTimezone)

	return &awssched.UpdateScheduleInput{
		Name:                       aws.String(cfg.Name),
		Description:                aws.String(cfg.Description),
		GroupName:                  aws.String(groupName),
		ScheduleExpression:         aws.String(expr),
		ScheduleExpressionTimezone: aws.String(timezone),
		FlexibleTimeWindow:         ftw,
		Target:                     target,
	}, nil
}

func (s *sched) buildCommon(cfg ScheduleConfig) (*types.FlexibleTimeWindow, string, *types.Target, error) {
	expr := fmt.Sprintf("at(%s)", cfg.ScheduleTime.UTC().Format("2006-01-02T15:04:05"))

	var ftw *types.FlexibleTimeWindow
	if cfg.FlexibleTimeWindowMinutes <= 0 {
		ftw = &types.FlexibleTimeWindow{Mode: types.FlexibleTimeWindowModeOff}
	} else {
		ftw = &types.FlexibleTimeWindow{
			Mode:                   types.FlexibleTimeWindowModeFlexible,
			MaximumWindowInMinutes: aws.Int32(int32(cfg.FlexibleTimeWindowMinutes)),
		}
	}

	target, err := s.buildTarget(cfg.Target)
	if err != nil {
		return nil, "", nil, err
	}
	return ftw, expr, target, nil
}

func (s *sched) buildTarget(t ScheduleTarget) (*types.Target, error) {
	switch target := t.(type) {
	case *LambdaTarget:
		return s.buildLambdaTarget(target)
	default:
		return nil, fmt.Errorf("scheduler: unsupported target type: %T", t)
	}
}

func (s *sched) buildLambdaTarget(t *LambdaTarget) (*types.Target, error) {
	payloadJSON, err := json.Marshal(t.Payload)
	if err != nil {
		return nil, fmt.Errorf("scheduler: failed to marshal lambda payload: %w", err)
	}

	maxAttempts := defaultMaxAttempts
	maxAgeSecs := defaultMaxAgeSecs
	if t.RetryPolicy != nil {
		if t.RetryPolicy.MaxAttempts > 0 {
			maxAttempts = t.RetryPolicy.MaxAttempts
		}
		if t.RetryPolicy.MaxEventAgeSecs > 0 {
			maxAgeSecs = t.RetryPolicy.MaxEventAgeSecs
		}
	}

	return &types.Target{
		Arn:     aws.String(t.LambdaArn),
		RoleArn: aws.String(s.roleArn),
		Input:   aws.String(string(payloadJSON)),
		RetryPolicy: &types.RetryPolicy{
			MaximumRetryAttempts:     aws.Int32(maxAttempts),
			MaximumEventAgeInSeconds: aws.Int32(maxAgeSecs),
		},
	}, nil
}

// --- helpers ---

func (s *sched) fail(span trace.Span, err error) error {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}

func (s *sched) logInfo(ctx context.Context, step, msg string, kv ...any) {
	if s.logger != nil {
		s.logger.Info(ctx, step, msg, kv...)
	}
}

func (s *sched) logError(ctx context.Context, step, msg string, kv ...any) {
	if s.logger != nil {
		s.logger.Error(ctx, step, msg, kv...)
	}
}

func validate(cfg ScheduleConfig) error {
	if cfg.Name == "" {
		return fmt.Errorf("scheduler: name is required")
	}
	if cfg.ScheduleTime.IsZero() {
		return fmt.Errorf("scheduler: schedule time is required")
	}
	if cfg.ScheduleTime.Before(time.Now()) {
		return fmt.Errorf("scheduler: schedule time %s is in the past", cfg.ScheduleTime.Format(time.RFC3339))
	}
	if cfg.Target == nil {
		return fmt.Errorf("scheduler: target is required")
	}
	switch t := cfg.Target.(type) {
	case *LambdaTarget:
		if t.LambdaArn == "" {
			return fmt.Errorf("scheduler: lambda ARN is required")
		}
	default:
		return fmt.Errorf("scheduler: unsupported target type: %T", t)
	}
	return nil
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}


