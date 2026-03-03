package dto

type GetCardOutput struct {
	ID          string
	Title       string
	Description string
	URL         string
	ListID      string
	Due         string
	Labels      []string
	Members     []string
}
