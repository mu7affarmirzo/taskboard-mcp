package entity

type Label struct {
	id    string
	name  string
	color string
}

func NewLabel(id, name, color string) *Label {
	return &Label{id: id, name: name, color: color}
}

func (l *Label) ID() string    { return l.id }
func (l *Label) Name() string  { return l.name }
func (l *Label) Color() string { return l.color }
