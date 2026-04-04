package timeutil

import (
	"testing"
	"time"
)

var fixedTime = time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)

func TestParseDate(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		got, err := ParseDate("")
		if err != nil || got != nil {
			t.Errorf("ParseDate(\"\") = %v, %v; want nil, nil", got, err)
		}
	})
	t.Run("valid date", func(t *testing.T) {
		got, err := ParseDate("2024-03-15")
		if err != nil || got == nil {
			t.Fatalf("ParseDate(\"2024-03-15\") err=%v", err)
		}
		if got.Year() != 2024 || got.Month() != 3 || got.Day() != 15 {
			t.Errorf("parsed date = %v, want 2024-03-15", got)
		}
	})
	t.Run("invalid format returns error", func(t *testing.T) {
		_, err := ParseDate("15-03-2024")
		if err == nil {
			t.Error("expected error for invalid format")
		}
	})
}

func TestParseDateTime(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		got, err := ParseDateTime("")
		if err != nil || got != nil {
			t.Errorf("ParseDateTime(\"\") = %v, %v; want nil, nil", got, err)
		}
	})
	t.Run("valid RFC3339", func(t *testing.T) {
		got, err := ParseDateTime("2024-03-15T10:30:00Z")
		if err != nil || got == nil {
			t.Fatalf("ParseDateTime err=%v", err)
		}
		if got.Year() != 2024 || got.Month() != 3 || got.Day() != 15 {
			t.Errorf("parsed = %v, want 2024-03-15", got)
		}
	})
	t.Run("invalid format returns error", func(t *testing.T) {
		_, err := ParseDateTime("2024-03-15")
		if err == nil {
			t.Error("expected error for invalid RFC3339")
		}
	})
}

func TestIsZeroOrNil(t *testing.T) {
	if !IsZeroOrNil(nil) {
		t.Error("IsZeroOrNil(nil) should be true")
	}
	zero := time.Time{}
	if !IsZeroOrNil(&zero) {
		t.Error("IsZeroOrNil(&zero) should be true")
	}
	if IsZeroOrNil(&fixedTime) {
		t.Error("IsZeroOrNil(fixedTime) should be false")
	}
}

func TestAfter(t *testing.T) {
	future := fixedTime.Add(time.Hour)
	past := fixedTime.Add(-time.Hour)

	if After(nil, fixedTime) {
		t.Error("After(nil, ...) should be false")
	}
	if !After(&future, fixedTime) {
		t.Error("After(future, fixed) should be true")
	}
	if After(&past, fixedTime) {
		t.Error("After(past, fixed) should be false")
	}
}

func TestBefore(t *testing.T) {
	future := fixedTime.Add(time.Hour)
	past := fixedTime.Add(-time.Hour)

	if Before(nil, fixedTime) {
		t.Error("Before(nil, ...) should be false")
	}
	if !Before(&past, fixedTime) {
		t.Error("Before(past, fixed) should be true")
	}
	if Before(&future, fixedTime) {
		t.Error("Before(future, fixed) should be false")
	}
}

func TestStartOfDay(t *testing.T) {
	got := StartOfDay(fixedTime)
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Errorf("StartOfDay = %v, want midnight", got)
	}
	if got.Year() != 2024 || got.Month() != 3 || got.Day() != 15 {
		t.Errorf("StartOfDay changed the date: %v", got)
	}
}

func TestEndOfDay(t *testing.T) {
	got := EndOfDay(fixedTime)
	if got.Hour() != 23 || got.Minute() != 59 || got.Second() != 59 {
		t.Errorf("EndOfDay = %v, want 23:59:59", got)
	}
	if got.Nanosecond() != int(time.Second-time.Nanosecond) {
		t.Errorf("EndOfDay nanoseconds = %d", got.Nanosecond())
	}
	if got.Year() != 2024 || got.Month() != 3 || got.Day() != 15 {
		t.Errorf("EndOfDay changed the date: %v", got)
	}
}

func TestStartEndOfDayRange(t *testing.T) {
	start := StartOfDay(fixedTime)
	end := EndOfDay(fixedTime)
	if !end.After(start) {
		t.Error("EndOfDay should be after StartOfDay")
	}
	if !fixedTime.After(start) || !fixedTime.Before(end) {
		t.Error("fixedTime should be within [StartOfDay, EndOfDay]")
	}
}
