package entity

type Board struct {
	id   string
	name string
}

func NewBoard(id, name string) *Board { return &Board{id: id, name: name} }
func (b *Board) ID() string           { return b.id }
func (b *Board) Name() string         { return b.name }

type BoardList struct {
	id   string
	name string
}

func NewBoardList(id, name string) *BoardList { return &BoardList{id: id, name: name} }
func (l *BoardList) ID() string               { return l.id }
func (l *BoardList) Name() string             { return l.name }
