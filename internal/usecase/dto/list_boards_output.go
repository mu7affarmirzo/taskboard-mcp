package dto

type ListBoardsOutput struct {
	Boards []BoardItem
}

type BoardItem struct {
	ID   string
	Name string
}
