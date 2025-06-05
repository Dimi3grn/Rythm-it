package common

// PaginationParams représente les paramètres de pagination
type PaginationParams struct {
	Page    int `json:"page" validate:"required,min=1"`
	PerPage int `json:"per_page" validate:"required,min=1,max=100"`
}

// ValidatePagination valide et normalise les paramètres de pagination
func ValidatePagination(params *PaginationParams) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}
}
