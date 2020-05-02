package storage

type Backend interface {
	Save(Message) (uint64, error)
	Get(uint64) (Message, error)
}

type Message interface {
}
