package dto

type ListCardsOutput struct {
	Cards []CardItem
}

type CardItem struct {
	ID     string
	Title  string
	URL    string
	ListID string
}
