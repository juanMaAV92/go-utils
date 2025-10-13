package pagination

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func ExtractParams(c echo.Context, defaultPage, defaultLimit int) (page, limit int) {
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	var err error

	if pageStr == "" {
		page = defaultPage
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return defaultPage, defaultLimit
		}
	}

	if limitStr == "" {
		limit = defaultLimit
	} else {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			return defaultPage, defaultLimit
		}
	}

	return page, limit
}

func BuildPagination(total int64, page, limit int) Pagination {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return Pagination{
		TotalPages: totalPages,
		TotalItems: int(total),
		Page:       page,
		Limit:      limit,
	}
}
