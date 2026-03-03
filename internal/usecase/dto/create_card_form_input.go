package dto

type CreateCardFormInput struct {
	TelegramID  int64
	ListID      string
	Title       string
	Description string
	DueDate     string
	LabelIDs    []string
	MemberIDs   []string
}
