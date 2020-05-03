package storage

type Message struct {
	From      string
	Channel   string
	Text      string
	Timestamp int64
}
