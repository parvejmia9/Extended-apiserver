package server

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type GenericServer struct {
	cfg Config
}

func NewGenericServer(cfg Config) *GenericServer {
	return &GenericServer{cfg: cfg}
}
func (s GenericServer) tlsConfig() *tls.Config {
	if s.cfg.CertFile == "" || s.cfg.KeyFile == "" {
		log.Fatalln("missing certificates, please provide it")
	}
	/*
		Ref:
		 - https://venilnoronha.io/a-step-by-step-guide-to-mtls-in-go
		 - https://blog.cloudflare.com/exposing-go-on-the-internet/
		 - http://www.hydrogen18.com/blog/your-own-pki-tls-golang.html
	*/
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
		SessionTicketsDisabled:   false,

		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		ClientAuth: tls.VerifyClientCertIfGiven,
		NextProtos: []string{"h2", "http/1.1"},
	}
	caCertPool := x509.NewCertPool()
	for _, cacFile := range s.cfg.CACertFiles {
		caCert, err := ioutil.ReadFile(cacFile)
		if err != nil {
			log.Fatalln(err)
		}
		caCertPool.AppendCertsFromPEM(caCert)
	}
	tlsConfig.ClientCAs = caCertPool
	tlsConfig.BuildNameToCertificate()
	return tlsConfig
}

func (s GenericServer) ListenAndServe(mux http.Handler) {
	log.Printf("Listening on %s\n", s.cfg.Address)
	srv := &http.Server{
		Addr:         s.cfg.Address,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	srv.TLSConfig = s.tlsConfig()
	log.Fatal(srv.ListenAndServeTLS(s.cfg.CertFile, s.cfg.KeyFile))
}
