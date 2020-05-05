package auth

import "channels/storage"

// TokenAuth is the auth Plugin which verify client with token.
type TokenAuth struct {
	Store storage.TokenBackend
}

// Authenticate returns a zero value Caller and nil (allow).
func (t *TokenAuth) Authenticate(user, token string) (*Caller, error) {
	data, err := t.Store.GetToken(token)
	if err != nil {
		return nil, err
	}
	return &Caller{
		Name: data.User,
		Caps: data.Scope,
	}, nil
}
