package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/gorilla/mux"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"net/http"
	"time"
)

func homeLink(w http.ResponseWriter, r *http.Request) {
	certRaw := "-----BEGIN CERTIFICATE-----\nMIIEFjCCAv6gAwIBAgIUUxC2WLk3GW3w8M5a2KwGbfq\n-----END CERTIFICATE-----"
	keyRaw := "-----BEGIN RSA PRIVATE KEY-----\nMIIEowsu22wlZ\n-----END RSA PRIVATE KEY-----"
	caRaw := "-----BEGIN CERTIFICATE-----\nMIIDzjCCAragAwIBAgIUadlZvb5y9TFfsbsyem9gh51SUNRCem\n-----END CERTIFICATE-----"

	certPair, err := tls.X509KeyPair([]byte(certRaw), []byte(keyRaw))
	if err != nil {
		return
	}

	certPoolCA := x509.NewCertPool()
	block, _ := pem.Decode([]byte(caRaw))

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return
	}

	certPoolCA.AddCert(cert)

	ctx := context.Background()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"etcd-stg-1.company:2379", "etcd-stg-2.company:2379", "etcd-stg-3.company:2379"},
		TLS: &tls.Config{
			RootCAs:      certPoolCA,
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{certPair},
		},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		// handle error!
	}
	defer cli.Close()

	lease, err := cli.Lease.Grant(ctx, 15)
	if err != nil {
		println("lease error")
	}

	leaseResponse := &LeaseResponse{LeaseID: int64(lease.ID)}
	jsonResponse, jsonError := json.Marshal(leaseResponse)

	if err != nil {
		println(jsonError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
	log.Fatal(http.ListenAndServe(":8080", router))
}

type LeaseResponse struct {
	LeaseID int64
}
