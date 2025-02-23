package subprober

import (
	"context"
	"testing"

	"github.com/BreakOnCrash/opendast/dns/client"
)

func TestProber(t *testing.T) {
	p := New(&Config{
		Dict: "../../tests/demo-dict.txt",
	}, client.NewClient(&client.Config{}))

	t.Log(p.Probe(context.TODO(), "example.com"))
}
