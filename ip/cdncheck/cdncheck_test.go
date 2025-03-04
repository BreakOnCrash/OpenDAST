package cdncheck

import (
	"fmt"
	"net"
	"testing"

	"github.com/projectdiscovery/cdncheck"
)

func TestCDNChekc(t *testing.T) {
	client := cdncheck.New()
	fmt.Println(client.Check(net.ParseIP("173.245.48.12")))

	fmt.Println(client.CheckDomainWithFallback("blog.imipy.com"))
}
