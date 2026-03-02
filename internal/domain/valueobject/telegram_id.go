package valueobject

type TelegramID int64

func NewTelegramID(id int64) TelegramID { return TelegramID(id) }
func (t TelegramID) Int64() int64       { return int64(t) }
