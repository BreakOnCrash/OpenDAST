package mitm

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/google/martian/v3"
	"github.com/google/martian/v3/mitm"
)

var defaultTimeout = 5 * time.Second

func RunDemo(certFile, keyFile, parentProxy string) {
	skipTLSVerify := true

	proxy := martian.NewProxy()
	proxy.SetRoundTripper(&http.Transport{
		MaxIdleConns:          100,
		TLSHandshakeTimeout:   defaultTimeout,
		ExpectContinueTimeout: defaultTimeout,
		ResponseHeaderTimeout: defaultTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
		},
	})

	if parentProxy != "" {
		proxyURL, err := url.Parse(parentProxy)
		if err != nil {
			log.Fatal(err)
		}
		proxy.SetDownstreamProxy(proxyURL)
	}

	// config mitm cert file
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	x509c, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		log.Fatal(err)
	}
	tlscnf, err := mitm.NewConfig(x509c, cert.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	tlscnf.SkipTLSVerify(skipTLSVerify)
	proxy.SetMITM(tlscnf)

	// set request, response modifier
	proxy.SetRequestModifier(martian.RequestModifierFunc(
		func(req *http.Request) error {
			log.Printf("[mitm] modify request - method: %s url: %s", req.Method, req.URL.String())
			return nil
		}))
	proxy.SetResponseModifier(martian.ResponseModifierFunc(
		func(res *http.Response) error {
			log.Printf("[mitm] modify response - method: %s url: %s status: %d", res.Request.Method, res.Request.URL.String(), res.StatusCode)
			return nil
		}))

	// start proxy server
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
	defer listener.Close()

	log.Printf("proxy server listen on %s", listener.Addr().String())
	if err := proxy.Serve(listener); err != nil {
		log.Fatalf("proxy serve error: %v", err)
	}
}
