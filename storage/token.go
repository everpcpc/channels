package storage

import (
	"fmt"
	"time"
)

type TokenBackend interface {
	GetToken(string) (*TokenData, error)
	AddToken(string, *TokenData) error
	DeleteTokens(...string) error
	ListTokens() (map[string]*TokenData, error)
}

type TokenData struct {
	User      string
	Scope     []string
	Note      string
	CreatedAt int64
}

func (t *TokenData) String() string {

	return fmt.Sprintf(
		"<token:%s%s%s(%s)>",
		t.User, t.Scope, t.Note, time.Unix(0, t.CreatedAt))
}
