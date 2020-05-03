package irc

// Config contains all of the configuration settings required to bring up a
// local irc server.
type Config struct {
	Name    string
	Network string
	Port    int
	SSLPort int

	PingFrequency  int
	PongMaxLatency int

	SSLCertificate SSLCertificate
}

// SSLCertificate contains the paths to the private key and certificate files to
// be used in SSL connections.
type SSLCertificate struct {
	KeyFile  string
	CertFile string
}

// // setConfigDefaults fills in the default values of the Config if no value is
// // specified for a field.
// func setConfigDefaults(cfg Config) Config {
// 	if cfg.PingFrequency == 0 {
// 		cfg.PingFrequency = 30
// 	}

// 	if cfg.PongMaxLatency == 0 {
// 		cfg.PongMaxLatency = 5
// 	}

// 	if cfg.SSLPort == 0 {
// 		cfg.SSLPort = 6697
// 	}

// 	return cfg
// }
