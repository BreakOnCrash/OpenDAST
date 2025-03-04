package geoip

import (
	"net"
	"testing"
)

func TestNewGeoIP(t *testing.T) {
	geo, err := NewGeoIP()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(geo.Find(net.ParseIP("115.192.37.128"), ""))
}
