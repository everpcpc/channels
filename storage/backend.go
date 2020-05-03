package storage

import "fmt"

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
