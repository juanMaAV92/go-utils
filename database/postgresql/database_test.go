package postgresql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		dst       interface{}
		wantErr   bool
		errString string
	}{
		{
			name:      "nil context",
			ctx:       nil,
			dst:       &struct{ Name string }{},
			wantErr:   true,
			errString: errContextRequired,
		},
		{
			name:      "nil destination",
			ctx:       context.Background(),
			dst:       nil,
			wantErr:   true,
			errString: errDestinationRequired,
		},
		{
			name:      "not a pointer destination",
			ctx:       context.Background(),
			dst:       struct{ Name string }{},
			wantErr:   true,
			errString: errDestinationMustBePointer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{}
			err := db.Create(tt.ctx, tt.dst)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errString, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		model     interface{}
		updates   map[string]interface{}
		wantErr   bool
		errString string
	}{
		{
			name:      "nil context",
			ctx:       nil,
			model:     &struct{ Name string }{},
			updates:   map[string]interface{}{"name": "test"},
			wantErr:   true,
			errString: errContextRequired,
		},
		{
			name:      "nil model",
			ctx:       context.Background(),
			model:     nil,
			updates:   map[string]interface{}{"name": "test"},
			wantErr:   true,
			errString: errModelRequired,
		},
		{
			name:      "nil updates",
			ctx:       context.Background(),
			model:     &struct{ Name string }{},
			updates:   nil,
			wantErr:   true,
			errString: errUpdatesRequired,
		},
		{
			name:      "empty updates",
			ctx:       context.Background(),
			model:     &struct{ Name string }{},
			updates:   map[string]interface{}{},
			wantErr:   true,
			errString: errUpdatesRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{}
			rowsAffected, err := db.Update(tt.ctx, tt.model, tt.updates, nil)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errString, err.Error())
				assert.Equal(t, int64(0), rowsAffected)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindOne(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		model     interface{}
		condition interface{}
		wantErr   bool
		errString string
	}{
		{
			name:      "nil context",
			ctx:       nil,
			model:     &struct{ Name string }{},
			condition: "name = 'test'",
			wantErr:   true,
			errString: errContextRequired,
		},
		{
			name:      "nil destination",
			ctx:       context.Background(),
			model:     nil,
			condition: "name = 'test'",
			wantErr:   true,
			errString: errDestinationRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{}
			found, err := db.FindOne(tt.ctx, tt.model, tt.condition, nil)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errString, err.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, found)
			}
		})
	}
}

func TestFindMany(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		model     interface{}
		condition interface{}
		wantErr   bool
		errString string
	}{
		{
			name:      "nil context",
			ctx:       nil,
			model:     &struct{ Name string }{},
			condition: "name = 'test'",
			wantErr:   true,
			errString: errContextRequired,
		},
		{
			name:      "nil destination",
			ctx:       context.Background(),
			model:     nil,
			condition: "name = 'test'",
			wantErr:   true,
			errString: errDestinationRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{}
			err := db.FindMany(tt.ctx, tt.model, tt.condition, &QueryOptions{
				Pagination: &PaginationOptions{
					Page:  1,
					Limit: 10,
				},
			})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errString, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		model     interface{}
		condition interface{}
		wantErr   bool
		errString string
	}{
		{
			name:      "nil context",
			ctx:       nil,
			model:     &struct{ Name string }{},
			condition: "name = 'test'",
			wantErr:   true,
			errString: errContextRequired,
		},
		{
			name:      "nil model",
			ctx:       context.Background(),
			model:     nil,
			condition: "name = 'test'",
			wantErr:   true,
			errString: errModelRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{}
			count, err := db.Count(tt.ctx, tt.model, tt.condition)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errString, err.Error())
				assert.Equal(t, int64(0), count)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindWithJoins(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		result    interface{}
		config    JoinConfig
		wantErr   bool
		errString string
	}{
		{
			name:   "nil context",
			ctx:    nil,
			result: &[]struct{ Name string }{},
			config: JoinConfig{
				BaseTable: "users",
				Joins: []JoinClause{
					{Type: "INNER", Table: "organizations", On: "users.id = organizations.user_id"},
				},
			},
			wantErr:   true,
			errString: errContextRequired,
		},
		{
			name:   "nil destination",
			ctx:    context.Background(),
			result: nil,
			config: JoinConfig{
				BaseTable: "users",
				Joins: []JoinClause{
					{Type: "INNER", Table: "organizations", On: "users.id = organizations.user_id"},
				},
			},
			wantErr:   true,
			errString: errDestinationRequired,
		},
		{
			name:   "empty base table",
			ctx:    context.Background(),
			result: &[]struct{ Name string }{},
			config: JoinConfig{
				BaseTable: "",
				Joins:     []JoinClause{},
			},
			wantErr:   true,
			errString: errBaseTableRequired,
		},
		{
			name:   "empty joins",
			ctx:    context.Background(),
			result: &[]struct{ Name string }{},
			config: JoinConfig{
				BaseTable: "users",
				Joins:     []JoinClause{},
			},
			wantErr:   true,
			errString: errJoinsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{}
			err := db.FindWithJoins(tt.ctx, tt.result, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errString, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRawQuery(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		result    interface{}
		query     string
		wantErr   bool
		errString string
	}{
		{
			name:      "nil context",
			ctx:       nil,
			result:    &[]struct{ Name string }{},
			query:     "SELECT * FROM users",
			wantErr:   true,
			errString: errContextRequired,
		},
		{
			name:      "empty query",
			ctx:       context.Background(),
			result:    &[]struct{ Name string }{},
			query:     "",
			wantErr:   true,
			errString: errQueryRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &Database{}
			result, err := db.ExecuteRawQuery(tt.ctx, tt.result, tt.query)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errString, err.Error())
				assert.Equal(t, QueryResult{}, result) // Should be empty result on error
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
