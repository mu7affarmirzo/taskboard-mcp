package dto

type ListLabelsOutput struct {
	Labels []LabelItem
}

type LabelItem struct {
	ID    string
	Name  string
	Color string
}
