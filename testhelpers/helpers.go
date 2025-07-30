package testhelpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

func PrepareContextFormTestCase(s *echo.Echo, test HttpTestCase) (echo.Context, *httptest.ResponseRecorder) {
	body, _ := json.Marshal(test.RequestBody)
	if str, ok := test.RequestBody.(string); ok {
		body = []byte(str)
	}

	url := prepareQueryParams(test.Request.Url, test.Request.QueryParam)

	req := httptest.NewRequest(test.Request.Method, url, strings.NewReader(string(body)))

	prepareHeaders(req, test.Request.Header)

	rec := httptest.NewRecorder()
	ctx := s.NewContext(req, rec)

	preparePathParams(ctx, test.Request.PathParam)

	return ctx, rec
}

func prepareQueryParams(baseURL string, queryParams []TestQueryParam) string {
	query := url.Values{}
	for _, param := range queryParams {
		query.Add(param.Name, param.Value)
	}
	if len(query) > 0 {
		return baseURL + "?" + query.Encode()
	}
	return baseURL
}

func prepareHeaders(req *http.Request, headers map[string]string) {
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func preparePathParams(ctx echo.Context, pathParams []TestPathParam) {
	if len(pathParams) > 0 {
		names := make([]string, len(pathParams))
		values := make([]string, len(pathParams))
		for i, param := range pathParams {
			names[i] = param.Name
			values[i] = param.Value
		}
		ctx.SetParamNames(names...)
		ctx.SetParamValues(values...)
	}
}

func ToJSONString(v interface{}) *string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	jsonStr := string(jsonBytes)
	return &jsonStr
}
