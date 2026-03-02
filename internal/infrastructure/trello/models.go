package trello

type CreateCardRequest struct {
	Name        string
	Description string
	ListID      string
	Due         string
	LabelIDs    []string
	Position    string
}

type CardResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ShortURL string `json:"shortUrl"`
	URL      string `json:"url"`
}

type BoardResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ListResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LabelResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}
