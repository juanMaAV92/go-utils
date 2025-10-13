package pagination

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestExtractPaginationParams(t *testing.T) {
	tests := []struct {
		name         string
		query        url.Values
		defaultPage  int
		defaultLimit int
		expectPage   int
		expectLimit  int
		expectErr    bool
	}{
		{
			name:         "No params, use defaults",
			query:        url.Values{},
			defaultPage:  2,
			defaultLimit: 20,
			expectPage:   2,
			expectLimit:  20,
			expectErr:    false,
		},
		{
			name:         "Valid params",
			query:        url.Values{"page": {"3"}, "limit": {"15"}},
			defaultPage:  1,
			defaultLimit: 10,
			expectPage:   3,
			expectLimit:  15,
			expectErr:    false,
		},
		{
			name:         "Invalid page",
			query:        url.Values{"page": {"abc"}},
			defaultPage:  1,
			defaultLimit: 10,
			expectPage:   1,
			expectLimit:  10,
			expectErr:    true,
		},
		{
			name:         "Invalid limit",
			query:        url.Values{"limit": {"-5"}},
			defaultPage:  1,
			defaultLimit: 10,
			expectPage:   1,
			expectLimit:  10,
			expectErr:    true,
		},
		{
			name:         "Page zero",
			query:        url.Values{"page": {"0"}},
			defaultPage:  1,
			defaultLimit: 10,
			expectPage:   1,
			expectLimit:  10,
			expectErr:    true,
		},
		{
			name:         "Limit zero",
			query:        url.Values{"limit": {"0"}},
			defaultPage:  1,
			defaultLimit: 10,
			expectPage:   1,
			expectLimit:  10,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.query.Encode(), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			page, limit := ExtractParams(c, tt.defaultPage, tt.defaultLimit)
			if tt.expectErr {
				assert.Equal(t, tt.expectPage, page)
				assert.Equal(t, tt.expectLimit, limit)
			} else {
				assert.Equal(t, tt.expectPage, page)
				assert.Equal(t, tt.expectLimit, limit)
			}
		})
	}
}

func TestBuildPagination(t *testing.T) {
	tests := []struct {
		name               string
		total              int64
		page               int
		limit              int
		expectedPagination Pagination
	}{
		{
			name:  "normal pagination",
			total: 100,
			page:  1,
			limit: 10,
			expectedPagination: Pagination{
				TotalPages: 10,
				TotalItems: 100,
				Page:       1,
				Limit:      10,
			},
		},
		{
			name:  "partial last page",
			total: 25,
			page:  3,
			limit: 10,
			expectedPagination: Pagination{
				TotalPages: 3,
				TotalItems: 25,
				Page:       3,
				Limit:      10,
			},
		},
		{
			name:  "empty result",
			total: 0,
			page:  1,
			limit: 10,
			expectedPagination: Pagination{
				TotalPages: 0,
				TotalItems: 0,
				Page:       1,
				Limit:      10,
			},
		},
		{
			name:  "single item",
			total: 1,
			page:  1,
			limit: 10,
			expectedPagination: Pagination{
				TotalPages: 1,
				TotalItems: 1,
				Page:       1,
				Limit:      10,
			},
		},
		{
			name:  "exact page boundary",
			total: 20,
			page:  2,
			limit: 10,
			expectedPagination: Pagination{
				TotalPages: 2,
				TotalItems: 20,
				Page:       2,
				Limit:      10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPagination(tt.total, tt.page, tt.limit)
			assert.Equal(t, tt.expectedPagination, result)
		})
	}
}
