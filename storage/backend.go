package storage

type Message struct {
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"` // TODO: support specific target
	Channel string `json:"channel,omitempty"`
	Text    string `json:"text,omitempty"`
}

type Backend interface {
	Save(Message) error
	PullLoop(chan Message)
	Subscribe(string) error
	UnSubscribe(string) error
}
