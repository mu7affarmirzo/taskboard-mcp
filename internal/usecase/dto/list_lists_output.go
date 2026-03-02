package dto

type ListListsOutput struct {
	Lists []ListItem
}

type ListItem struct {
	ID   string
	Name string
}
