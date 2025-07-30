package errors

import (
	"net/http"
)

var UnsupportedMediaType = ErrorResponse{
	Code:     StatusUnsupportedMediaTypeCode,
	Messages: []string{"Unsupported media type"},
	HttpCode: http.StatusUnsupportedMediaType,
}

var StatusNotFound = ErrorResponse{
	Code:     StatusNotFoundCode,
	Messages: []string{"Not found"},
	HttpCode: http.StatusNotFound,
}

var Unauthorized = ErrorResponse{
	Code:     StatusUnauthorizedCode,
	Messages: []string{"Unauthorized"},
	HttpCode: http.StatusUnauthorized,
}

var Forbidden = ErrorResponse{
	Code:     StatusForbiddenCode,
	Messages: []string{"Forbidden"},
	HttpCode: http.StatusForbidden,
}

var MethodNotAllowed = ErrorResponse{
	Code:     StatusMethodNotAllowedCode,
	Messages: []string{"Method not allowed"},
	HttpCode: http.StatusMethodNotAllowed,
}

var RequestEntityTooLarge = ErrorResponse{
	Code:     StatusRequestEntityTooLargeCode,
	Messages: []string{"Request entity too large"},
	HttpCode: http.StatusRequestEntityTooLarge,
}

var TooManyRequests = ErrorResponse{
	Code:     StatusTooManyRequestsCode,
	Messages: []string{"Too many requests"},
	HttpCode: http.StatusTooManyRequests,
}

var BadRequest = ErrorResponse{
	Code:     StatusBadRequestCode,
	Messages: []string{"Bad request"},
	HttpCode: http.StatusBadRequest,
}

var BadGateway = ErrorResponse{
	Code:     StatusBadGatewayCode,
	Messages: []string{"Bad gateway"},
	HttpCode: http.StatusBadGateway,
}

var InternalServerError = ErrorResponse{
	Code:     StatusInternalServerErrorCode,
	Messages: []string{"Internal error"},
	HttpCode: http.StatusInternalServerError,
}

var RequestTimeout = ErrorResponse{
	Code:     StatusRequestTimeoutCode,
	Messages: []string{"Request timeout"},
	HttpCode: http.StatusRequestTimeout,
}

var ServiceUnavailable = ErrorResponse{
	Code:     StatusServiceUnavailableCode,
	Messages: []string{"Service unavailable"},
	HttpCode: http.StatusServiceUnavailable,
}
