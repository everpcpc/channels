package auth

// Caller represents an client making a request. Callers are determined by
// the Plugin Authenticate method. The default Plugin (Anonymous) returns a zero
// value Caller.
type Caller struct {
	// Name of the caller, whether human (username) or machine (app name). The
	// name is user-defined and only used by Spin Cycle for logging and setting
	// proto.Request.User.
	Name string

	// Roles are user-defined role names, like "admin" or "engineer". Rolls
	// are matched against request ACL roles in specs, which are also user-defined.
	// Roles are case-sensitive and not modified in any way.
	Roles []string
}

// Plugin represents the auth plugin. Every request is authenticated and authorized.
// The default Plugin (AllowAll) allows everything: all callers, requests, and ops.
type Plugin interface {
	// Authenticate caller. If allowed, return nil.
	// Access is denied on any error.
	//
	// The returned Caller is user-defined. The most important field is Roles
	// which will be matched against request ACL roles from the specs.
	Authenticate(user, pass string) (Caller, error)
}
