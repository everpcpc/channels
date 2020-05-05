package storage

type TokenBackend interface {
	GetToken(string) (*TokenData, error)
	AddToken(string, *TokenData) error
	DeleteToken(string) error
}

type TokenData struct {
	User      string
	Scope     []string
	CreatedAt int64
	Note      string
}
