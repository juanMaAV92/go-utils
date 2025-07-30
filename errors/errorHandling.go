package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CustomHTTPErrorHandler(err error, ctx echo.Context) {

	if ctx.Response().Committed {
		return
	}
	var customError *ErrorResponse
	if errors.As(err, &customError) {
		ctx.JSON(customError.ErrorHTTPCode(), ErrorResponse{
			Code:     customError.ErrorCode(),
			Messages: customError.ErrorMessages(),
			HttpCode: customError.ErrorHTTPCode(),
		})
		return
	}

	var errorEcho *echo.HTTPError
	if errors.As(err, &errorEcho) {
		ctx.JSON(errorEcho.Code, echoErrorToErrorResponse(errorEcho))
		return
	}

	ctx.JSON(http.StatusInternalServerError, InternalServerError)
}

var mapEchoErrorsToErrorResponse = map[int]*ErrorResponse{
	http.StatusUnsupportedMediaType:  &UnsupportedMediaType,
	http.StatusNotFound:              &StatusNotFound,
	http.StatusUnauthorized:          &Unauthorized,
	http.StatusForbidden:             &Forbidden,
	http.StatusMethodNotAllowed:      &MethodNotAllowed,
	http.StatusRequestEntityTooLarge: &RequestEntityTooLarge,
	http.StatusTooManyRequests:       &TooManyRequests,
	http.StatusBadRequest:            &BadRequest,
	http.StatusBadGateway:            &BadGateway,
	http.StatusInternalServerError:   &InternalServerError,
	http.StatusRequestTimeout:        &RequestTimeout,
	http.StatusServiceUnavailable:    &ServiceUnavailable,
}

func echoErrorToErrorResponse(httpError *echo.HTTPError) ErrorResponse {
	var errorResponse = mapEchoErrorsToErrorResponse[httpError.Code]
	if errorResponse == nil {
		errorResponse = &ErrorResponse{
			Code:     fmt.Sprint(httpError.Code),
			Messages: []string{fmt.Sprintf("%+v", httpError.Message)},
		}
	}

	return *errorResponse
}
