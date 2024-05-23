package main

import (
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/parvejmia9/extended-apiserver/lib/certstore"
	"github.com/parvejmia9/extended-apiserver/lib/server"
	"github.com/spf13/afero"
	"k8s.io/client-go/util/cert"
	"log"
	"net"
	"net/http"
)

func main() {
	var proxy = false
	flag.BoolVar(&proxy, "receive-proxy-request", proxy, "receive sent request from apiserver")
	flag.Parse()
	fs := afero.NewOsFs()
	store, err := certstore.NewCertStore(fs, certstore.CertDir)
	if err != nil {
		log.Fatal(err)
	}
	err = store.InitCA("database")
	if err != nil {
		log.Fatalln(err)
	}
	serverCert, serverKey, err := store.NewServerCertPair(cert.AltNames{
		IPs: []net.IP{net.ParseIP("127.0.0.2")},
	})
	if err != nil {
		log.Fatal(err)
	}
	err = store.Write("tls", serverCert, serverKey)
	if err != nil {
		log.Fatal(err)
	}
	clientCert, clientKey, err := store.NewClientCertPair(cert.AltNames{
		DNSNames: []string{"parvejmia9-database"},
	})
	if err != nil {
		log.Fatal(err)
	}
	err = store.Write("parvejmia9-database", clientCert, clientKey)
	if err != nil {
		log.Fatal(err)
	}
	apiServerStore, err := certstore.NewCertStore(fs, certstore.CertDir)
	if err != nil {
		log.Fatal(err)
	}
	if proxy {
		err = apiServerStore.LoadCA("apiserver")
		if err != nil {
			log.Fatal(err)
		}
	}
	rhCACertPool := x509.NewCertPool()
	rhStore, err := certstore.NewCertStore(fs, certstore.CertDir)
	if err != nil {
		log.Fatal(err)
	}
	if proxy {
		err = rhStore.LoadCA("requestheader")
		if err != nil {
			log.Fatal(err)
		}
		rhCACertPool.AppendCertsFromPEM(rhStore.CACertBytes())
	}
	cfg := server.Config{
		Address:     "127.0.0.2:8989",
		CACertFiles: []string{},
		CertFile:    store.CertFile("tls"),
		KeyFile:     store.KeyFile("tls"),
	}
	if proxy {
		cfg.CACertFiles = append(cfg.CACertFiles, apiServerStore.CertFile("ca"))
		cfg.CACertFiles = append(cfg.CACertFiles, rhStore.CertFile("ca"))
	}

	srv := server.NewGenericServer(cfg)

	r := mux.NewRouter()

	r.HandleFunc("/database/{resource}", func(w http.ResponseWriter, req *http.Request) {
		user := "system:anonymous"
		src := "-"
		if len(req.TLS.PeerCertificates) > 0 {
			opts := x509.VerifyOptions{
				Roots:     rhCACertPool,
				KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			}
			if _, err := req.TLS.PeerCertificates[0].Verify(opts); err != nil {
				user = req.TLS.PeerCertificates[0].Subject.CommonName
				src = "Client-Cert-CN"
			} else {
				user = req.Header.Get("X-Remote-User")
				src = "X-Remote-User"
			}
		}
		vars := mux.Vars(req)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Resource: %v requested by user [%s]=%s\n", vars["resource"], src, user)
	})
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "OK")
	})
	srv.ListenAndServe(r)

}
