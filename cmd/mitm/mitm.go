package main

import "github.com/BreakOnCrash/opendast/mitm"

func main() {
	certFile := "./ca.pem"
	keyFile := "./ca.key"

	mitm.RunDemo(certFile, keyFile, "")
}
