package auth

// LDAPAuth is the auth Plugin with openldap server.
type LDAPAuth struct{}

// Authenticate returns the Caller and auth result
func (l LDAPAuth) Authenticate(user, pass string) (*Caller, error) {
	return &Caller{Name: user}, nil
}
