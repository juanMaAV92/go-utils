package env_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/juanmaAV/go-utils/env"
)

func TestGetEnv(t *testing.T) {
	t.Run("returns value", func(t *testing.T) {
		t.Setenv("TEST_KEY", "hello")
		if got := env.GetEnv("TEST_KEY"); got != "hello" {
			t.Errorf("got %q, want \"hello\"", got)
		}
	})

	t.Run("panics when unset", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		env.GetEnv("UNSET_VAR_XYZ")
	})

	t.Run("panics when blank", func(t *testing.T) {
		t.Setenv("BLANK_VAR", "   ")
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for blank value")
			}
		}()
		env.GetEnv("BLANK_VAR")
	})
}

func TestGetEnvWithDefault(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		t.Setenv("K", "v")
		if got := env.GetEnvWithDefault("K", "default"); got != "v" {
			t.Errorf("got %q, want \"v\"", got)
		}
	})

	t.Run("returns default when unset", func(t *testing.T) {
		_ = os.Unsetenv("K")
		if got := env.GetEnvWithDefault("K", "default"); got != "default" {
			t.Errorf("got %q, want \"default\"", got)
		}
	})
}

func TestGetEnvironment(t *testing.T) {
	t.Run("returns ENVIRONMENT value", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "production")
		if got := env.GetEnvironment(); got != "production" {
			t.Errorf("got %q, want \"production\"", got)
		}
	})

	t.Run("defaults to local", func(t *testing.T) {
		_ = os.Unsetenv("ENVIRONMENT")
		if got := env.GetEnvironment(); got != "local" {
			t.Errorf("got %q, want \"local\"", got)
		}
	})
}

func TestGetEnvAsIntWithDefault(t *testing.T) {
	t.Run("returns parsed int", func(t *testing.T) {
		t.Setenv("N", "42")
		if got := env.GetEnvAsIntWithDefault("N", 0); got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})

	t.Run("returns default when unset", func(t *testing.T) {
		_ = os.Unsetenv("N")
		if got := env.GetEnvAsIntWithDefault("N", 99); got != 99 {
			t.Errorf("got %d, want 99", got)
		}
	})

	t.Run("panics on invalid int", func(t *testing.T) {
		t.Setenv("N", "nope")
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		env.GetEnvAsIntWithDefault("N", 0)
	})
}

func TestGetEnvAsDurationWithDefault(t *testing.T) {
	t.Run("returns parsed duration", func(t *testing.T) {
		t.Setenv("D", "5m")
		if got := env.GetEnvAsDurationWithDefault("D", time.Second); got != 5*time.Minute {
			t.Errorf("got %v, want 5m", got)
		}
	})

	t.Run("returns default when unset", func(t *testing.T) {
		_ = os.Unsetenv("D")
		if got := env.GetEnvAsDurationWithDefault("D", 30*time.Second); got != 30*time.Second {
			t.Errorf("got %v, want 30s", got)
		}
	})

	t.Run("panics on invalid duration", func(t *testing.T) {
		t.Setenv("D", "invalid")
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		env.GetEnvAsDurationWithDefault("D", time.Second)
	})
}

func TestMustHave(t *testing.T) {
	t.Run("no panic when all set", func(t *testing.T) {
		t.Setenv("A", "1")
		t.Setenv("B", "2")
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()
		env.MustHave("A", "B")
	})

	t.Run("panics listing all missing vars", func(t *testing.T) {
		_ = os.Unsetenv("MISSING_X")
		_ = os.Unsetenv("MISSING_Y")
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
			msg := fmt.Sprintf("%v", r)
			if !strings.Contains(msg, "MISSING_X") || !strings.Contains(msg, "MISSING_Y") {
				t.Errorf("panic message should list all missing vars, got: %s", msg)
			}
		}()
		env.MustHave("MISSING_X", "MISSING_Y")
	})

	t.Run("panics only for missing ones", func(t *testing.T) {
		t.Setenv("PRESENT", "ok")
		_ = os.Unsetenv("ABSENT")
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
			msg := fmt.Sprintf("%v", r)
			if strings.Contains(msg, "PRESENT") {
				t.Errorf("should not list PRESENT in missing vars: %s", msg)
			}
			if !strings.Contains(msg, "ABSENT") {
				t.Errorf("should list ABSENT in missing vars: %s", msg)
			}
		}()
		env.MustHave("PRESENT", "ABSENT")
	})
}

func TestGetEnvAsSliceWithDefault(t *testing.T) {
	t.Run("splits by separator", func(t *testing.T) {
		t.Setenv("S", "a,b,c")
		got := env.GetEnvAsSliceWithDefault("S", ",", nil)
		if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
			t.Errorf("got %v, want [a b c]", got)
		}
	})

	t.Run("trims whitespace from elements", func(t *testing.T) {
		t.Setenv("S", "a , b , c")
		got := env.GetEnvAsSliceWithDefault("S", ",", nil)
		for _, v := range got {
			if v != strings.TrimSpace(v) {
				t.Errorf("element %q was not trimmed", v)
			}
		}
	})

	t.Run("returns default when unset", func(t *testing.T) {
		_ = os.Unsetenv("S")
		def := []string{"x"}
		got := env.GetEnvAsSliceWithDefault("S", ",", def)
		if len(got) != 1 || got[0] != "x" {
			t.Errorf("got %v, want default", got)
		}
	})

	t.Run("excludes empty elements", func(t *testing.T) {
		t.Setenv("S", "a,,b")
		got := env.GetEnvAsSliceWithDefault("S", ",", nil)
		if len(got) != 2 {
			t.Errorf("got %v, want [a b] (no empty elements)", got)
		}
	})
}

func TestGetEnvAsBoolWithDefault(t *testing.T) {
	cases := []struct {
		val  string
		want bool
	}{
		{"true", true},
		{"1", true},
		{"t", true},
		{"false", false},
		{"0", false},
		{"f", false},
	}
	for _, c := range cases {
		t.Run(c.val, func(t *testing.T) {
			t.Setenv("B", c.val)
			if got := env.GetEnvAsBoolWithDefault("B", !c.want); got != c.want {
				t.Errorf("val %q: got %v, want %v", c.val, got, c.want)
			}
		})
	}

	t.Run("returns default when unset", func(t *testing.T) {
		_ = os.Unsetenv("B")
		if got := env.GetEnvAsBoolWithDefault("B", true); got != true {
			t.Error("expected default true")
		}
	})

	t.Run("panics on invalid bool", func(t *testing.T) {
		t.Setenv("B", "yes")
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		env.GetEnvAsBoolWithDefault("B", false)
	})
}
