package storage

type Backend interface {
	Save(Message) error
	PullLoop(chan Message)
	Subscribe(string) error
	UnSubscribe(string) error
}
