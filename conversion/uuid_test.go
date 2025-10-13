package conversion

import (
	"testing"

	"github.com/google/uuid"
)

func TestUUIDToString(t *testing.T) {
	u := uuid.New()

	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"nil input", nil, ""},
		{"string input", "test-string", "test-string"},
		{"uuid input", u, u.String()},
		{"unknown type", 123, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UUIDToString(tt.input)
			if got != tt.want {
				t.Errorf("UUIDToString(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestToUUID(t *testing.T) {
	u := uuid.New()

	tests := []struct {
		name    string
		input   interface{}
		want    uuid.UUID
		wantErr bool
	}{
		{"nil input", nil, uuid.Nil, false},
		{"uuid input", u, u, false},
		{"valid string", u.String(), u, false},
		{"invalid string", "not-a-uuid", uuid.Nil, true},
		{"unknown type", 123, uuid.Nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToUUID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToUUID(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ToUUID(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
