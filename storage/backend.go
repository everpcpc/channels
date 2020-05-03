package storage

import "fmt"

type Message struct {
	From    string `json:"from,omitempty"`
	Channel string `json:"channel,omitempty"`
	Text    string `json:"text,omitempty"`
}

type Backend interface {
	Save(Message) error
	PullLoop(chan Message)
	Subscribe(string) error
	UnSubscribe(string) error
}

func New(store, addr string) (Backend, error) {
	switch store {
	case "redis":
		return NewRedisBackend(addr)
	}
	return nil, fmt.Errorf("backend %s not supported", store)
}
