package pagination

type Pagination struct {
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
}
