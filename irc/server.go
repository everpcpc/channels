package irc

import (
	"crypto/tls"
	"fmt"
	"net"

	"mcdc/state"
)

// RunServer starts the IRC server. This method will not return.
func RunServer(cfg Config) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logf(fatal, "Could not create server: %v", err)
	}

	var lnSSL net.Listener
	certFile := cfg.SSLCertificate.CertFile
	keyFile := cfg.SSLCertificate.KeyFile
	if keyFile != "" && certFile != "" {
		getCertificate := func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err == nil {
				return &cert, nil
			} else {
				logf(warn, "Could not create TLS server: %v", err)
				return nil, err
			}
		}

		sslCfg := &tls.Config{GetCertificate: getCertificate}

		lnSSL, err = tls.Listen("tcp", fmt.Sprintf(":%d", cfg.SSLPort), sslCfg)
		if err != nil {
			logf(fatal, "Could not create TLS server: %v", err)
		}
	}
	name := "ircd"
	s := make(chan state.State, 1)
	s <- state.New(name)

	if lnSSL != nil {
		go acceptLoop(cfg, lnSSL, s)
	}
	acceptLoop(cfg, ln, s)
}

func acceptLoop(cfg Config, listener net.Listener, s chan state.State) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			logf(warn, "Could not accept new connection: ", err)
			continue
		}

		c := newConnection(cfg, conn, newFreshHandler(s))
		go c.loop()
	}
}
