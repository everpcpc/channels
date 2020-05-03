package auth

// Anonymous is the default Plugin which allows all requests (no auth).
type Anonymous struct{}

// Authenticate returns a zero value Caller and nil (allow).
func (a Anonymous) Authenticate(user, pass string) (*Caller, error) {
	return &Caller{Name: user}, nil
}
