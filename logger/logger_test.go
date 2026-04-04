package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func newTestLogger(buf *bytes.Buffer, opts ...Option) Logger {
	return newWithWriter(buf, "test", opts...)
}

func parseLog(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("failed to parse log output %q: %v", buf.String(), err)
	}
	return m
}

func TestInfo_FlatFields(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)
	log.Info(context.Background(), "user.create", "user created", "userID", "123", "role", "admin")

	m := parseLog(t, &buf)

	if m["step"] != "user.create" {
		t.Errorf("step = %v, want \"user.create\"", m["step"])
	}
	if m["message"] != "user created" {
		t.Errorf("message = %v, want \"user created\"", m["message"])
	}
	// fields must be flat, not nested under "attributes"
	if m["userID"] != "123" {
		t.Errorf("userID = %v, want \"123\"", m["userID"])
	}
	if m["role"] != "admin" {
		t.Errorf("role = %v, want \"admin\"", m["role"])
	}
	if _, nested := m["attributes"]; nested {
		t.Error("fields must be flat, not nested under \"attributes\"")
	}
}

func TestInfo_NoArgs(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)
	log.Info(context.Background(), "health.check", "ok")

	m := parseLog(t, &buf)
	if m["step"] != "health.check" {
		t.Errorf("step = %v", m["step"])
	}
}

func TestInfo_OddArgs_DoesNotPanic(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)
	// odd number of args — last key has no value, should not panic
	log.Info(context.Background(), "test", "msg", "orphan_key")
}

func TestWithLevel_FiltersDebug(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf, WithLevel(InfoLevel))
	log.Debug(context.Background(), "test", "should not appear")

	if buf.Len() > 0 {
		t.Errorf("debug log should be filtered at InfoLevel, got: %s", buf.String())
	}
}

func TestWithLevel_AllowsDebug(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf, WithLevel(DebugLevel))
	log.Debug(context.Background(), "test", "should appear")

	if buf.Len() == 0 {
		t.Error("debug log should appear at DebugLevel")
	}
}

func TestError_Level(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)
	log.Error(context.Background(), "db.query", "connection failed", "err", "timeout")

	m := parseLog(t, &buf)
	if m["level"] != "ERROR" {
		t.Errorf("level = %v, want \"ERROR\"", m["level"])
	}
	if m["err"] != "timeout" {
		t.Errorf("err field = %v, want \"timeout\"", m["err"])
	}
}

func TestServiceField(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)
	log.Info(context.Background(), "test", "msg")

	m := parseLog(t, &buf)
	if m["service"] != "test" {
		t.Errorf("service = %v, want \"test\"", m["service"])
	}
}

func TestMessageKey_IsMessage(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)
	log.Info(context.Background(), "test", "hello world")

	m := parseLog(t, &buf)
	if _, hasMsg := m["msg"]; hasMsg {
		t.Error("key should be \"message\", not \"msg\"")
	}
	if m["message"] != "hello world" {
		t.Errorf("message = %v, want \"hello world\"", m["message"])
	}
}
