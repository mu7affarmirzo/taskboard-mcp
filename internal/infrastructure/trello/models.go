package trello

type CreateCardRequest struct {
	Name        string
	Description string
	ListID      string
	Due         string
	LabelIDs    []string
	MemberIDs   []string
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

type MemberResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

type UpdateCardRequest struct {
	Name      *string
	Desc      *string
	IDList    *string
	Due       *string
	IDLabels  *string
	IDMembers *string
}

type CardDetailResponse struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Desc     string           `json:"desc"`
	ShortURL string           `json:"shortUrl"`
	URL      string           `json:"url"`
	IDList   string           `json:"idList"`
	Due      string           `json:"due"`
	Closed   bool             `json:"closed"`
	Labels   []LabelResponse  `json:"labels"`
	Members  []MemberResponse `json:"members"`
}

type CommentResponse struct {
	ID   string            `json:"id"`
	Data CommentDataResponse `json:"data"`
}

type CommentDataResponse struct {
	Text string `json:"text"`
}

type SearchResponse struct {
	Cards []CardDetailResponse `json:"cards"`
}
