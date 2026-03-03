package dto

type ActionOutput struct {
	Action  string
	Message string
	CardURL string
	Items   []ActionItem
}

type ActionItem struct {
	ID    string
	Name  string
	URL   string
	Extra string
}
