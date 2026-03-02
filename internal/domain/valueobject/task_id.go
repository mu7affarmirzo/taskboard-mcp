package valueobject

type TaskID string

func NewTaskID(id string) TaskID { return TaskID(id) }
func (t TaskID) String() string  { return string(t) }
