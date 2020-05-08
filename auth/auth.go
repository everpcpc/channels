package auth

// Plugin represents the auth plugin. Every request is authenticated and authorized.
// The default Plugin (AllowAll) allows everything: all callers, requests, and ops.
type Plugin interface {
	// Authenticate caller. If allowed, return nil.
	// Access is denied on any error.
	//
	// The returned Caller is user-defined. The most important field is Roles
	// which will be matched against request ACL roles from the specs.
	Authenticate(string, string) (*Caller, error)
}

// Caller represents an client making a request. Callers are determined by
// the Plugin Authenticate method. The default Plugin (Anonymous) returns a zero
// value Caller.
type Caller struct {
	// Name of the caller, whether human (username) or machine (app name).
	Name string

	// Roles are user-defined role names, like "admin" or "engineer".
	// Rolls are used by user api to check read/write permissions.
	// Roles are case-sensitive and not modified in any way.
	Roles []string

	// Caps are channel names, user names etc.
	// Caps are used by webhook api for auth scope
	// Caps are case-sensitive and not modified in any way.
	Caps []string
}

func (c *Caller) IsCapable(target string) bool {
	if target == "" {
		return false
	}

	switch target[0] {
	case '#':
		for _, cap := range c.Caps {
			if cap == "#" || cap == target {
				return true
			}
		}
	case '@':
		for _, cap := range c.Caps {
			if cap == "@" || cap == target {
				return true
			}
		}
	}

	return false
}
