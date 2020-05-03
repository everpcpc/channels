package storage

import "fmt"

type Backend interface {
	Save(Message) error
}

func New(store, addr string) (Backend, error) {
	switch store {
	case "redis":
		return NewRedisBackend(addr)
	}
	return nil, fmt.Errorf("backend %s not supported", store)
}
