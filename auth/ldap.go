package auth

import (
	"errors"
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

// LDAPAuth is the auth Plugin with openldap server.
type LDAPAuth struct {
	// ldap://example.com:389 or ldaps://example.com:636
	URL string

	BindDN   string
	BindPass string

	SearchFilter string
	SearchBase   string

	AttrUsername string
	AttrMemberOf string
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
		err = fmt.Errorf("server bind error: %v", err)
		return
	}

	req := ldap.NewSearchRequest(
		l.SearchBase, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 1, 5, false,
		fmt.Sprintf(l.SearchFilter, user),
		[]string{l.AttrUsername, l.AttrMemberOf}, nil,
	)
	res, err := ld.Search(req)
	if err != nil {
		err = fmt.Errorf("search error: %v", err)
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
			var g *ldap.DN
			for _, dn := range attr.Values {
				g, err = ldap.ParseDN(dn)
				if err != nil {
					err = fmt.Errorf("get user group error: %v", err)
					return
				}
				for _, attrs := range g.RDNs {
					for _, attr := range attrs.Attributes {
						if attr.Type == "cn" {
							userGroups = append(userGroups, attr.Value)
						}
					}
				}
			}
		default:
		}
	}
	err = ld.Bind(res.Entries[0].DN, pass)
	if err != nil {
		err = fmt.Errorf("login error: %v", err)
		return
	}

	return &Caller{Name: user, Roles: userGroups}, nil
}
