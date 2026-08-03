package params

import (
	"net/http"
	"strconv"
)

// OrderParams embeds standard Pagination fields and adds order-specific filters
type OrderParams struct {
	MemberID int
	Pagination
}

// OrderParamsFromRequest parses query params including member_id and pagination
func OrderParamsFromRequest(r *http.Request) OrderParams {
	p := FromRequest(r) // Parse limit, offset, sort_by, order, search (q)

	var memberID int
	if memberStr := r.URL.Query().Get("member_id"); memberStr != "" {
		if id, err := strconv.Atoi(memberStr); err == nil && id > 0 {
			memberID = id
		}
	}

	return OrderParams{
		MemberID:   memberID,
		Pagination: p,
	}
}
