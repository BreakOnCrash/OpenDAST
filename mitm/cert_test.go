package main

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/google/martian/v3/mitm"
)

func TestGenCA(t *testing.T) {
	x509c, priv, err := mitm.NewAuthority("zznq.mitm", "ZZNQ MITM", 10*365*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	certOut, err := os.Create("./ca.pem")
	if err != nil {
		t.Fatal(err)
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: x509c.Raw})

	keyOut, err := os.Create("./ca.key")
	if err != nil {
		t.Fatal(err)
	}
	defer keyOut.Close()

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
}
