package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		key       string
		value     interface{}
		wantErr   bool
		errString string
	}{
		{
			name:      "nil context",
			ctx:       nil,
			key:       "test-key",
			value:     "test-value",
			wantErr:   true,
			errString: "context cannot be nil",
		},
		{
			name:      "empty key",
			ctx:       context.Background(),
			key:       "",
			value:     "test-value",
			wantErr:   true,
			errString: "key cannot be empty",
		},
		{
			name:      "nil value",
			ctx:       context.Background(),
			key:       "test-key",
			value:     nil,
			wantErr:   true,
			errString: "value cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cache{}
			err := c.Set(tt.ctx, tt.key, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		key       string
		dest      interface{}
		wantErr   bool
		errString string
	}{
		{
			name:      "nil context",
			ctx:       nil,
			key:       "test-key",
			dest:      &struct{}{},
			wantErr:   true,
			errString: "context cannot be nil",
		},
		{
			name:      "empty key",
			ctx:       context.Background(),
			key:       "",
			dest:      &struct{}{},
			wantErr:   true,
			errString: "key cannot be empty",
		},
		{
			name:      "nil destination",
			ctx:       context.Background(),
			key:       "test-key",
			dest:      nil,
			wantErr:   true,
			errString: "value cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cache{}
			_, err := c.Get(tt.ctx, tt.key, tt.dest)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		key       string
		wantErr   bool
		errString string
	}{
		{
			name:      "nil context",
			ctx:       nil,
			key:       "test-key",
			wantErr:   true,
			errString: "context cannot be nil",
		},
		{
			name:      "empty key",
			ctx:       context.Background(),
			key:       "",
			wantErr:   true,
			errString: "key cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cache{}
			_, err := c.Exists(tt.ctx, tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
