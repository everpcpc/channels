package irc

// Config contains all of the configuration settings required to bring up a
// local irc server.
type Config struct {
	Name    string
	Port    int
	SSLPort int

	PingFrequency  int
	PongMaxLatency int

	SSLCertificate SSLCertificate

	Auth string
}

// SSLCertificate contains the paths to the private key and certificate files to
// be used in SSL connections.
type SSLCertificate struct {
	KeyFile  string
	CertFile string
}
