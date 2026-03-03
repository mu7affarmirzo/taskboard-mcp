package dto

type UpdateCardInput struct {
	Title       *string
	Description *string
	ListID      *string
	Due         *string
	LabelIDs    *string
	MemberIDs   *string
}
