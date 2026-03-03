package dto

type GetUserSettingsOutput struct {
	TrelloConnected bool
	DefaultBoardID  string
	DefaultListID   string
}
