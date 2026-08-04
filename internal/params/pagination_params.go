package params

import (
	"net/http"
	"strconv"
	"strings"
)

// Pagination holds shared query fields
type Pagination struct {
	Limit  int
	Offset int
	SortBy string
	Order  string
	Search string
}

// FromRequest extracts standard pagination parameters with defaults
func FromRequest(r *http.Request) Pagination {
	query := r.URL.Query()

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit <= 0 {
		limit = 20 // Default page size
	}

	offset, err := strconv.Atoi(query.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	order := strings.ToUpper(query.Get("order"))
	if order != "DESC" {
		order = "ASC"
	}

	return Pagination{
		Limit:  limit,
		Offset: offset,
		SortBy: query.Get("sort_by"),
		Order:  order,
		Search: query.Get("q"),
	}
}
