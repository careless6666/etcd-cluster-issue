package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/coreos/etcd/etcdserver/api/v3lock/v3lockpb"
	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

func main() {
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		go TakeLock(ctx)
	}

	time.Sleep(100 * time.Second)
}

func TakeLock(ctx context.Context) {
	println("hello world")

	certRaw := "-----BEGIN CERTIFICATE-----\nMIIEFjCCAv6gAwIBAgIUUxC2WLk3Guop7XAztxPwRqsgqLlUKNoBnXlXG57o5Qj0uFf/6ooEw+BW3w8M5a2KwGbfq\n-----END CERTIFICATE-----"
	keyRaw := "-----BEGIN RSA PRIVATE KEY-----\Qsu22wlZ\n-----END RSA PRIVATE KEY-----"
	caRaw := "-----BEGIN CERTIFICATE-----\nMIIDzjCCAragAwIBAgIUadlZvb5y9TFfsbsy
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

	tlsCredentials := &tls.Config{
		RootCAs:      certPoolCA,
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{certPair},
	}

	cc1, err := grpc.Dial("dns:///etcd-stg-1.company:2379", grpc.WithTransportCredentials(credentials.NewTLS(tlsCredentials)))
	if err != nil {
		println("cannot dial server: ", err)
	}

	lockClient := v3lockpb.NewLockClient(cc1)

	var leaseClient = resty.New()

	lessResponse := &LeaseResponse{}
	lessResponseError := ""

	reponseText, err := leaseClient.
		EnableTrace().
		R().
		SetResult(&lessResponse).
		SetError(&lessResponseError).
		Get("http://localhost:8080")

	println(reponseText)

	lockRequest := &v3lockpb.LockRequest{
		Name:  []byte("/ic-me-daemon-global-sync/election"),
		Lease: lessResponse.LeaseID,
	}

	_, err = lockClient.Lock(ctx, lockRequest)
	if err != nil {
		println("cannot dial server: ", err.Error())
	} else {
		time.Sleep(300 * time.Millisecond)
	}

	println("exit")

}

type LeaseResponse struct {
	LeaseID int64
}
