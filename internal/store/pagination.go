package store

type paginationWindow struct {
	Page       int
	TotalPages int
	StartItem  int
	EndItem    int
	Offset     int
}

func paginate(totalResults, page, perPage int) paginationWindow {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 1
	}
	if totalResults == 0 {
		return paginationWindow{Page: 1, TotalPages: 1}
	}

	totalPages := (totalResults + perPage - 1) / perPage
	if page > totalPages {
		page = totalPages
	}
	startItem := ((page - 1) * perPage) + 1
	endItem := startItem + perPage - 1
	if endItem > totalResults {
		endItem = totalResults
	}

	return paginationWindow{
		Page:       page,
		TotalPages: totalPages,
		StartItem:  startItem,
		EndItem:    endItem,
		Offset:     (page - 1) * perPage,
	}
}
