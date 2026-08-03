package params

import (
	"net/http"
	"strconv"
)

// BookshelfParams embeds the standard Pagination fields
type BookshelfParams struct {
	LibraryID int
	Pagination
}

// BookParams embeds the standard Pagination fields
type BookParams struct {
	GenreID int
	Pagination
}

func BookParamsFromRequest(r *http.Request) BookParams {
	p := FromRequest(r) // Parse standard pagination fields (limit, offset, sort_by, order, search)

	var genreID int
	if genreStr := r.URL.Query().Get("genre_id"); genreStr != "" {
		if id, err := strconv.Atoi(genreStr); err == nil && id > 0 {
			genreID = id
		}
	}

	return BookParams{
		GenreID:    genreID,
		Pagination: p,
	}
}
