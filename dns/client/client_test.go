package client

import (
	"testing"

	"github.com/miekg/dns"
)

func TestParseResolvers(t *testing.T) {
	for _, rr := range ParseResolvers([]string{"udp:1.1.1.1:53", "8.8.8.8", "8.8.8.8:54", "udp:"}) {
		t.Log(rr.Proto())
		t.Log(rr.Addr())
	}
}

func TestClient(t *testing.T) {
	c := NewClient(&Config{
		Resolvers: []string{"2.3.3.3", "8.8.8.8"},
	})

	t.Log(c.Resolve("google.com"))

	d, err := c.QueryMultiple("google.com", []uint16{dns.TypeA, dns.TypeAAAA})
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range d {
		t.Log(r)
	}
}

func TestBaseDemo(t *testing.T) {
	c := dns.Client{Net: "udp"}                  // 创建客户端
	msg := &dns.Msg{}                            // 构造查询报文
	msg.SetQuestion("baidu.com.", dns.TypeA)     // 查询 baidu.com A 记录
	res, _, err := c.Exchange(msg, "8.8.8.8:53") // 向 8.8.8.8 域名服务器发送请求
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
