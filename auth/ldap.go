package auth

import (
	"errors"

	"github.com/go-ldap/ldap/v3"
)

// LDAPAuth is the auth Plugin with openldap server.
type LDAPAuth struct {
	URL      string
	BindDN   string
	BindPass string

	SearchFilter string
	SearchBase   []string

	AttrUsername string
	AttrMemberOf string
}

func NewLDAPAuth() *LDAPAuth {
	return &LDAPAuth{
		URL:          "ldap://example.com:389",
		BindDN:       "uid=test,dc=example,dc=com",
		BindPass:     "password",
		SearchFilter: "(uid=%s)",
		SearchBase:   []string{"ou=people,dc=example,dc=com"},
		AttrUsername: "uid",
		AttrMemberOf: "memberOf",
	}
}

// Authenticate returns the Caller and auth result
func (l *LDAPAuth) Authenticate(user, pass string) (c *Caller, err error) {
	ld, err := ldap.DialURL(l.URL)
	if err != nil {
		return
	}
	defer ld.Close()

	err = ld.Bind(l.BindDN, l.BindPass)
	if err != nil {
		return
	}

	req := ldap.NewSearchRequest(
		user, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 1, 5, false,
		l.SearchFilter, []string{l.AttrUsername, l.AttrMemberOf}, nil,
	)
	res, err := ld.Search(req)
	if err != nil {
		return
	}
	if len(res.Entries) == 0 {
		err = errors.New("user not found")
		return
	}

	var userGroups []string

	for _, attr := range res.Entries[0].Attributes {
		switch attr.Name {
		case l.AttrUsername:
			if len(attr.Values) != 1 {
				err = errors.New("username not unique")
			}
			if attr.Values[0] != user {
				err = errors.New("username mismatch")
			}
		case l.AttrMemberOf:
			userGroups = attr.Values
		default:
		}
	}
	err = ld.Bind(res.Entries[0].DN, pass)
	if err != nil {
		return
	}

	return &Caller{Name: user, Roles: userGroups}, nil
}
