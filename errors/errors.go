package errors

import (
	"fmt"
	"net/http"
)

// ErrorResponse is a structured HTTP error that implements the error interface.
// HttpCode is excluded from JSON — it is used to set the response status code.
type ErrorResponse struct {
	Code     string   `json:"code"`
	Messages []string `json:"messages,omitempty"`
	HttpCode int      `json:"-"`
}

// New creates a custom ErrorResponse.
func New(httpCode int, code string, messages []string) *ErrorResponse {
	return &ErrorResponse{
		HttpCode: httpCode,
		Code:     code,
		Messages: messages,
	}
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("HTTP_STATUS=%d CODE=%s MESSAGES=%v", e.HttpCode, e.Code, e.Messages)
}

func (e *ErrorResponse) ErrorCode() string     { return e.Code }
func (e *ErrorResponse) ErrorMessages() []string { return e.Messages }
func (e *ErrorResponse) ErrorHTTPCode() int    { return e.HttpCode }

// WithMessage returns a copy of e with Messages replaced by a single message.
func (e *ErrorResponse) WithMessage(msg string) *ErrorResponse {
	return &ErrorResponse{
		HttpCode: e.HttpCode,
		Code:     e.Code,
		Messages: []string{msg},
	}
}

// WithMessages returns a copy of e with Messages replaced.
func (e *ErrorResponse) WithMessages(msgs []string) *ErrorResponse {
	return &ErrorResponse{
		HttpCode: e.HttpCode,
		Code:     e.Code,
		Messages: msgs,
	}
}

// Predefined constructors — each call returns a fresh *ErrorResponse.

func ErrBadRequest(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusBadRequest, Code: StatusBadRequestCode, Messages: defaultMessages(messages, "bad request")}
}

func ErrUnauthorized(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusUnauthorized, Code: StatusUnauthorizedCode, Messages: defaultMessages(messages, "unauthorized")}
}

func ErrForbidden(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusForbidden, Code: StatusForbiddenCode, Messages: defaultMessages(messages, "forbidden")}
}

func ErrNotFound(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusNotFound, Code: StatusNotFoundCode, Messages: defaultMessages(messages, "not found")}
}

func ErrMethodNotAllowed(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusMethodNotAllowed, Code: StatusMethodNotAllowedCode, Messages: defaultMessages(messages, "method not allowed")}
}

func ErrRequestTimeout(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusRequestTimeout, Code: StatusRequestTimeoutCode, Messages: defaultMessages(messages, "request timeout")}
}

func ErrTooManyRequests(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusTooManyRequests, Code: StatusTooManyRequestsCode, Messages: defaultMessages(messages, "too many requests")}
}

func ErrRequestEntityTooLarge(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusRequestEntityTooLarge, Code: StatusRequestEntityTooLargeCode, Messages: defaultMessages(messages, "request entity too large")}
}

func ErrUnsupportedMediaType(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusUnsupportedMediaType, Code: StatusUnsupportedMediaTypeCode, Messages: defaultMessages(messages, "unsupported media type")}
}

func ErrInternalServer(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusInternalServerError, Code: StatusInternalServerErrorCode, Messages: defaultMessages(messages, "internal server error")}
}

func ErrBadGateway(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusBadGateway, Code: StatusBadGatewayCode, Messages: defaultMessages(messages, "bad gateway")}
}

func ErrServiceUnavailable(messages ...string) *ErrorResponse {
	return &ErrorResponse{HttpCode: http.StatusServiceUnavailable, Code: StatusServiceUnavailableCode, Messages: defaultMessages(messages, "service unavailable")}
}

func defaultMessages(provided []string, fallback string) []string {
	if len(provided) > 0 {
		return provided
	}
	return []string{fallback}
}
