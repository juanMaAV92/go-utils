package echoerr

import (
	stderrors "errors"
	"fmt"
	"net/http"

	"github.com/juanmaAV/go-utils/errors"
	"github.com/labstack/echo/v4"
)

// HTTPErrorHandler is an Echo HTTPErrorHandler that serializes both
// *errors.ErrorResponse and *echo.HTTPError into a consistent JSON body.
//
// Register with: e.HTTPErrorHandler = echoerr.HTTPErrorHandler
func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var appErr *errors.ErrorResponse
	if stderrors.As(err, &appErr) {
		_ = c.JSON(appErr.ErrorHTTPCode(), appErr)
		return
	}

	var echoErr *echo.HTTPError
	if stderrors.As(err, &echoErr) {
		resp := echoErrToResponse(echoErr)
		_ = c.JSON(echoErr.Code, resp)
		return
	}

	_ = c.JSON(http.StatusInternalServerError, errors.ErrInternalServer())
}

func echoErrToResponse(e *echo.HTTPError) *errors.ErrorResponse {
	switch e.Code {
	case http.StatusBadRequest:
		return errors.ErrBadRequest()
	case http.StatusUnauthorized:
		return errors.ErrUnauthorized()
	case http.StatusForbidden:
		return errors.ErrForbidden()
	case http.StatusNotFound:
		return errors.ErrNotFound()
	case http.StatusMethodNotAllowed:
		return errors.ErrMethodNotAllowed()
	case http.StatusRequestTimeout:
		return errors.ErrRequestTimeout()
	case http.StatusTooManyRequests:
		return errors.ErrTooManyRequests()
	case http.StatusRequestEntityTooLarge:
		return errors.ErrRequestEntityTooLarge()
	case http.StatusUnsupportedMediaType:
		return errors.ErrUnsupportedMediaType()
	case http.StatusInternalServerError:
		return errors.ErrInternalServer()
	case http.StatusBadGateway:
		return errors.ErrBadGateway()
	case http.StatusServiceUnavailable:
		return errors.ErrServiceUnavailable()
	default:
		return errors.New(e.Code, fmt.Sprintf("%d", e.Code), []string{fmt.Sprintf("%v", e.Message)})
	}
}
