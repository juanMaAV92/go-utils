package errors

// HTTP Status Error Codes
const (
	// 4xx Client Errors
	StatusBadRequestCode            = "BAD_REQUEST"
	StatusUnauthorizedCode          = "UNAUTHORIZED"
	StatusForbiddenCode             = "FORBIDDEN"
	StatusNotFoundCode              = "NOT_FOUND"
	StatusMethodNotAllowedCode      = "METHOD_NOT_ALLOWED"
	StatusRequestTimeoutCode        = "REQUEST_TIMEOUT"
	StatusRequestEntityTooLargeCode = "REQUEST_ENTITY_TOO_LARGE"
	StatusUnsupportedMediaTypeCode  = "UNSUPPORTED_MEDIA_TYPE"
	StatusTooManyRequestsCode       = "TOO_MANY_REQUESTS"

	// 5xx Server Errors
	StatusInternalServerErrorCode = "INTERNAL_ERROR"
	StatusBadGatewayCode          = "BAD_GATEWAY"
	StatusServiceUnavailableCode  = "SERVICE_UNAVAILABLE"
)

// Validation Error Codes
const (
	ValidationErrorCode = "VALIDATION_ERROR"
	InvalidRequestCode  = "INVALID_REQUEST"
)
