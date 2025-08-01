package testhelpers

import (
	"github.com/juanMaAV92/go-utils/errors"
	"github.com/labstack/echo/v4"
)

type HttpTestCase struct {
	TestName    string
	Request     TestRequest
	RequestBody interface{}
	MockFunc    func(s *echo.Echo, c echo.Context)
	Response    ExpectedResponse
	ExpectError *errors.ErrorResponse
}

type TestRequest struct {
	Method     string
	Url        string
	PathParam  []TestPathParam
	Header     map[string]string
	QueryParam []TestQueryParam
}

type TestPathParam struct {
	Name  string
	Value string
}

type TestQueryParam struct {
	Name  string
	Value string
}

type ExpectedResponse struct {
	Status int
	Body   *string
}
