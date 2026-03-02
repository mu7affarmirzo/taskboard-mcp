package entity

import "telegram-trello-bot/internal/domain/valueobject"

type User struct {
	telegramID   valueobject.TelegramID
	trelloToken  string
	defaultBoard string
	defaultList  string
	useLLM       bool
}

func NewUser(telegramID valueobject.TelegramID) *User {
	return &User{telegramID: telegramID, useLLM: true}
}

func (u *User) TelegramID() valueobject.TelegramID { return u.telegramID }
func (u *User) TrelloToken() string                { return u.trelloToken }
func (u *User) DefaultBoard() string               { return u.defaultBoard }
func (u *User) DefaultList() string                { return u.defaultList }
func (u *User) HasBoardConfigured() bool           { return u.defaultBoard != "" }
func (u *User) HasListConfigured() bool            { return u.defaultList != "" }
func (u *User) UseLLM() bool                       { return u.useLLM }

func (u *User) SetTrelloToken(token string)    { u.trelloToken = token }
func (u *User) SetDefaultBoard(boardID string) { u.defaultBoard = boardID }
func (u *User) SetDefaultList(listID string)   { u.defaultList = listID }
func (u *User) SetUseLLM(use bool)             { u.useLLM = use }
