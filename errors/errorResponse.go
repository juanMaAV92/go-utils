package errors

import "fmt"

type ErrorResponse struct {
	Code     string   `json:"code"`
	Messages []string `json:"messages,omitempty"`
	HttpCode int      `json:"-"`
	Internal error    `json:"-"`
}

func New(httpCode int, code string, messages []string) *ErrorResponse {
	return &ErrorResponse{
		Code:     code,
		Messages: messages,
		HttpCode: httpCode,
	}
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("HTTP_STATUS=%d--CODE=%s--MESSAGES=%v", e.HttpCode, e.Code, e.Messages)
}

func (e *ErrorResponse) ErrorCode() string {
	return e.Code
}

func (e *ErrorResponse) ErrorMessages() []string {
	return e.Messages
}

func (e *ErrorResponse) ErrorHTTPCode() int {
	return e.HttpCode
}
