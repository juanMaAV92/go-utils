package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	awssched "github.com/aws/aws-sdk-go-v2/service/scheduler"
	"go.opentelemetry.io/otel/trace/noop"
)

// --- mock ---

type mockScheduler struct {
	createFn func(ctx context.Context, params *awssched.CreateScheduleInput, optFns ...func(*awssched.Options)) (*awssched.CreateScheduleOutput, error)
	updateFn func(ctx context.Context, params *awssched.UpdateScheduleInput, optFns ...func(*awssched.Options)) (*awssched.UpdateScheduleOutput, error)
	deleteFn func(ctx context.Context, params *awssched.DeleteScheduleInput, optFns ...func(*awssched.Options)) (*awssched.DeleteScheduleOutput, error)
	getFn    func(ctx context.Context, params *awssched.GetScheduleInput, optFns ...func(*awssched.Options)) (*awssched.GetScheduleOutput, error)
}

func (m *mockScheduler) CreateSchedule(ctx context.Context, params *awssched.CreateScheduleInput, optFns ...func(*awssched.Options)) (*awssched.CreateScheduleOutput, error) {
	return m.createFn(ctx, params, optFns...)
}
func (m *mockScheduler) UpdateSchedule(ctx context.Context, params *awssched.UpdateScheduleInput, optFns ...func(*awssched.Options)) (*awssched.UpdateScheduleOutput, error) {
	return m.updateFn(ctx, params, optFns...)
}
func (m *mockScheduler) DeleteSchedule(ctx context.Context, params *awssched.DeleteScheduleInput, optFns ...func(*awssched.Options)) (*awssched.DeleteScheduleOutput, error) {
	return m.deleteFn(ctx, params, optFns...)
}
func (m *mockScheduler) GetSchedule(ctx context.Context, params *awssched.GetScheduleInput, optFns ...func(*awssched.Options)) (*awssched.GetScheduleOutput, error) {
	return m.getFn(ctx, params, optFns...)
}

func newTestSched(mock schedulerAPI) *sched {
	return &sched{
		client:  mock,
		logger:  nil,
		tracer:  noop.NewTracerProvider().Tracer("test"),
		roleArn: "arn:aws:iam::123456789012:role/test-role",
	}
}

func futureTime() time.Time {
	return time.Now().Add(1 * time.Hour)
}

func validLambdaConfig(name string) ScheduleConfig {
	return ScheduleConfig{
		Name:         name,
		ScheduleTime: futureTime(),
		Target: &LambdaTarget{
			LambdaArn: "arn:aws:lambda:us-east-1:123456789012:function:my-fn",
			Payload:   map[string]any{"key": "value"},
		},
	}
}

// --- CreateSchedule ---

func TestCreateSchedule_Success(t *testing.T) {
	expectedArn := "arn:aws:scheduler:us-east-1:123456789012:schedule/default/my-schedule"
	s := newTestSched(&mockScheduler{
		createFn: func(_ context.Context, params *awssched.CreateScheduleInput, _ ...func(*awssched.Options)) (*awssched.CreateScheduleOutput, error) {
			if *params.Name != "my-schedule" {
				t.Errorf("unexpected name: %s", *params.Name)
			}
			return &awssched.CreateScheduleOutput{ScheduleArn: &expectedArn}, nil
		},
	})

	result, err := s.CreateSchedule(context.Background(), validLambdaConfig("my-schedule"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ScheduleArn != expectedArn {
		t.Errorf("expected arn %s, got %s", expectedArn, result.ScheduleArn)
	}
	if result.ScheduleName != "my-schedule" {
		t.Errorf("expected name my-schedule, got %s", result.ScheduleName)
	}
}

func TestCreateSchedule_ValidationFails_EmptyName(t *testing.T) {
	s := newTestSched(&mockScheduler{})
	cfg := validLambdaConfig("")
	_, err := s.CreateSchedule(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreateSchedule_ValidationFails_PastTime(t *testing.T) {
	s := newTestSched(&mockScheduler{})
	cfg := validLambdaConfig("test")
	cfg.ScheduleTime = time.Now().Add(-1 * time.Hour)
	_, err := s.CreateSchedule(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for past time")
	}
}

func TestCreateSchedule_ValidationFails_NoTarget(t *testing.T) {
	s := newTestSched(&mockScheduler{})
	cfg := validLambdaConfig("test")
	cfg.Target = nil
	_, err := s.CreateSchedule(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for nil target")
	}
}

func TestCreateSchedule_ValidationFails_EmptyLambdaArn(t *testing.T) {
	s := newTestSched(&mockScheduler{})
	cfg := validLambdaConfig("test")
	cfg.Target = &LambdaTarget{LambdaArn: ""}
	_, err := s.CreateSchedule(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for empty lambda ARN")
	}
}

func TestCreateSchedule_ClientError(t *testing.T) {
	s := newTestSched(&mockScheduler{
		createFn: func(_ context.Context, _ *awssched.CreateScheduleInput, _ ...func(*awssched.Options)) (*awssched.CreateScheduleOutput, error) {
			return nil, errors.New("aws error")
		},
	})
	_, err := s.CreateSchedule(context.Background(), validLambdaConfig("my-schedule"))
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- UpdateSchedule ---

func TestUpdateSchedule_Success(t *testing.T) {
	expectedArn := "arn:aws:scheduler:us-east-1:123456789012:schedule/default/my-schedule"
	s := newTestSched(&mockScheduler{
		updateFn: func(_ context.Context, params *awssched.UpdateScheduleInput, _ ...func(*awssched.Options)) (*awssched.UpdateScheduleOutput, error) {
			return &awssched.UpdateScheduleOutput{ScheduleArn: &expectedArn}, nil
		},
	})

	result, err := s.UpdateSchedule(context.Background(), validLambdaConfig("my-schedule"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ScheduleArn != expectedArn {
		t.Errorf("expected arn %s, got %s", expectedArn, result.ScheduleArn)
	}
}

func TestUpdateSchedule_ClientError(t *testing.T) {
	s := newTestSched(&mockScheduler{
		updateFn: func(_ context.Context, _ *awssched.UpdateScheduleInput, _ ...func(*awssched.Options)) (*awssched.UpdateScheduleOutput, error) {
			return nil, errors.New("aws error")
		},
	})
	_, err := s.UpdateSchedule(context.Background(), validLambdaConfig("my-schedule"))
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- DeleteSchedule ---

func TestDeleteSchedule_Success(t *testing.T) {
	called := false
	s := newTestSched(&mockScheduler{
		deleteFn: func(_ context.Context, params *awssched.DeleteScheduleInput, _ ...func(*awssched.Options)) (*awssched.DeleteScheduleOutput, error) {
			called = true
			if *params.Name != "my-schedule" {
				t.Errorf("unexpected name: %s", *params.Name)
			}
			if *params.GroupName != "my-group" {
				t.Errorf("unexpected group: %s", *params.GroupName)
			}
			return &awssched.DeleteScheduleOutput{}, nil
		},
	})

	err := s.DeleteSchedule(context.Background(), "my-schedule", "my-group")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("delete was not called")
	}
}

func TestDeleteSchedule_DefaultGroup(t *testing.T) {
	s := newTestSched(&mockScheduler{
		deleteFn: func(_ context.Context, params *awssched.DeleteScheduleInput, _ ...func(*awssched.Options)) (*awssched.DeleteScheduleOutput, error) {
			if *params.GroupName != defaultGroupName {
				t.Errorf("expected default group, got %s", *params.GroupName)
			}
			return &awssched.DeleteScheduleOutput{}, nil
		},
	})

	err := s.DeleteSchedule(context.Background(), "my-schedule", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSchedule_EmptyName(t *testing.T) {
	s := newTestSched(&mockScheduler{})
	err := s.DeleteSchedule(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDeleteSchedule_ClientError(t *testing.T) {
	s := newTestSched(&mockScheduler{
		deleteFn: func(_ context.Context, _ *awssched.DeleteScheduleInput, _ ...func(*awssched.Options)) (*awssched.DeleteScheduleOutput, error) {
			return nil, errors.New("aws error")
		},
	})
	err := s.DeleteSchedule(context.Background(), "my-schedule", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- GetSchedule ---

func TestGetSchedule_Success(t *testing.T) {
	name := "my-schedule"
	arn := "arn:aws:scheduler:us-east-1:123456789012:schedule/default/my-schedule"
	group := "default"
	s := newTestSched(&mockScheduler{
		getFn: func(_ context.Context, params *awssched.GetScheduleInput, _ ...func(*awssched.Options)) (*awssched.GetScheduleOutput, error) {
			return &awssched.GetScheduleOutput{
				Name:      &name,
				Arn:       &arn,
				GroupName: &group,
			}, nil
		},
	})

	info, err := s.GetSchedule(context.Background(), "my-schedule", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != name {
		t.Errorf("expected name %s, got %s", name, info.Name)
	}
	if info.Arn != arn {
		t.Errorf("expected arn %s, got %s", arn, info.Arn)
	}
}

func TestGetSchedule_EmptyName(t *testing.T) {
	s := newTestSched(&mockScheduler{})
	_, err := s.GetSchedule(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestGetSchedule_ClientError(t *testing.T) {
	s := newTestSched(&mockScheduler{
		getFn: func(_ context.Context, _ *awssched.GetScheduleInput, _ ...func(*awssched.Options)) (*awssched.GetScheduleOutput, error) {
			return nil, errors.New("aws error")
		},
	})
	_, err := s.GetSchedule(context.Background(), "my-schedule", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- buildCommon / FlexibleTimeWindow ---

func TestBuildCommon_ExactTime(t *testing.T) {
	s := newTestSched(&mockScheduler{})
	cfg := validLambdaConfig("test")
	ftw, expr, target, err := s.buildCommon(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ftw.Mode != "OFF" {
		t.Errorf("expected OFF mode, got %s", ftw.Mode)
	}
	if expr == "" {
		t.Error("expected non-empty expression")
	}
	if target == nil {
		t.Error("expected non-nil target")
	}
}

func TestBuildCommon_FlexibleWindow(t *testing.T) {
	s := newTestSched(&mockScheduler{})
	cfg := validLambdaConfig("test")
	cfg.FlexibleTimeWindowMinutes = 15
	ftw, _, _, err := s.buildCommon(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ftw.Mode != "FLEXIBLE" {
		t.Errorf("expected FLEXIBLE mode, got %s", ftw.Mode)
	}
	if *ftw.MaximumWindowInMinutes != 15 {
		t.Errorf("expected 15 minutes, got %d", *ftw.MaximumWindowInMinutes)
	}
}

// --- RetryPolicy ---

func TestBuildLambdaTarget_DefaultRetryPolicy(t *testing.T) {
	s := newTestSched(&mockScheduler{})
	t2 := &LambdaTarget{
		LambdaArn: "arn:aws:lambda:us-east-1:123456789012:function:my-fn",
		Payload:   map[string]any{},
	}
	target, err := s.buildLambdaTarget(t2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *target.RetryPolicy.MaximumRetryAttempts != defaultMaxAttempts {
		t.Errorf("expected %d attempts, got %d", defaultMaxAttempts, *target.RetryPolicy.MaximumRetryAttempts)
	}
	if *target.RetryPolicy.MaximumEventAgeInSeconds != defaultMaxAgeSecs {
		t.Errorf("expected %d secs, got %d", defaultMaxAgeSecs, *target.RetryPolicy.MaximumEventAgeInSeconds)
	}
}

func TestBuildLambdaTarget_CustomRetryPolicy(t *testing.T) {
	s := newTestSched(&mockScheduler{})
	t2 := &LambdaTarget{
		LambdaArn:   "arn:aws:lambda:us-east-1:123456789012:function:my-fn",
		Payload:     map[string]any{},
		RetryPolicy: &RetryPolicy{MaxAttempts: 5, MaxEventAgeSecs: 600},
	}
	target, err := s.buildLambdaTarget(t2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *target.RetryPolicy.MaximumRetryAttempts != 5 {
		t.Errorf("expected 5 attempts, got %d", *target.RetryPolicy.MaximumRetryAttempts)
	}
	if *target.RetryPolicy.MaximumEventAgeInSeconds != 600 {
		t.Errorf("expected 600 secs, got %d", *target.RetryPolicy.MaximumEventAgeInSeconds)
	}
}

// --- NewWebhookSchedule ---

func TestNewWebhookSchedule(t *testing.T) {
	cfg := NewWebhookSchedule(
		"test", "desc", "grp",
		futureTime(),
		"arn:aws:lambda:us-east-1:123456789012:function:proxy",
		"https://example.com/webhook",
		"POST",
		map[string]any{"order_id": "123"},
	)
	if cfg.Name != "test" {
		t.Errorf("unexpected name: %s", cfg.Name)
	}
	lt, ok := cfg.Target.(*LambdaTarget)
	if !ok {
		t.Fatal("expected LambdaTarget")
	}
	if lt.Payload["url"] != "https://example.com/webhook" {
		t.Errorf("unexpected url: %v", lt.Payload["url"])
	}
	if lt.Payload["method"] != "POST" {
		t.Errorf("unexpected method: %v", lt.Payload["method"])
	}
}
