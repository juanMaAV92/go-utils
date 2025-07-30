package errors

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	httpCode := 404
	code := "NOT_FOUND"
	messages := []string{"Resource not found"}
	errResp := New(httpCode, code, messages)

	if errResp.Error() != "HTTP_STATUS=404--CODE=NOT_FOUND--MESSAGES=[Resource not found]" {
		t.Errorf("expected Error %s, got %s", "HTTP_STATUS=404--CODE=NOT_FOUND--MESSAGES=[Resource not found]", errResp.Error())
	}

	if errResp.ErrorCode() != code {
		t.Errorf("expected Code %s, got %s", code, errResp.Code)
	}

	if len(errResp.ErrorMessages()) != len(messages) {
		t.Errorf("expected Messages %v, got %v", messages, errResp.Messages)
	}
	for i, msg := range errResp.ErrorMessages() {
		if msg != messages[i] {
			t.Errorf("expected Message %s, got %s", messages[i], msg)
		}
	}

	if errResp.ErrorHTTPCode() != httpCode {
		t.Errorf("expected ErrorHTTPCode %d, got %d", httpCode, errResp.ErrorHTTPCode())
	}
}

func Test_EchoErrorToErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		inputError     *echo.HTTPError
		expectedOutput ErrorResponse
	}{
		{
			name: "Known error code",
			inputError: &echo.HTTPError{
				Code:    http.StatusNotFound,
				Message: "Resource not found",
			},
			expectedOutput: *mapEchoErrorsToErrorResponse[http.StatusNotFound],
		},
		{
			name: "Unknown error code",
			inputError: &echo.HTTPError{
				Code:    999,
				Message: "Unknown error",
			},
			expectedOutput: ErrorResponse{
				Code:     "999",
				Messages: []string{"Unknown error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := echoErrorToErrorResponse(tt.inputError)
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}
