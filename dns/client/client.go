package client

import (
	"errors"
	"net"
	"time"

	"github.com/miekg/dns"
)

type Config struct {
	Resolvers  []string `yaml:"resolvers" json:"resolvers"`
	Timeout    int      `yaml:"timeout" json:"timeout"`
	MaxRetries int      `yaml:"max-retries" json:"max-retries"`
}

type DNSRecord struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
	Value  string `json:"value"`
}

var (
	ErrMaxRetries = errors.New("could not resolve, max retries exceeded")
)

type Client struct {
	maxRetries int         // 最大重试次数
	resolvers  []Resolver  // DNS服务器
	udpClient  *dns.Client // udp连接
	tcpClient  *dns.Client // tcp连接
}

func NewClient(cfg *Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 // default timeout 5 seconds
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3 // default max retries 3
	}
	if len(cfg.Resolvers) == 0 {
		cfg.Resolvers = DefaultResolvers // default resolvers
	}

	timeout := time.Duration(cfg.Timeout) * time.Second
	return &Client{
		maxRetries: cfg.MaxRetries,
		resolvers:  ParseResolvers(cfg.Resolvers),
		udpClient: &dns.Client{
			Net:     "udp",
			Timeout: timeout,
			Dialer:  &net.Dialer{},
		},
		tcpClient: &dns.Client{
			Net:     "tcp",
			Timeout: timeout,
			Dialer:  &net.Dialer{},
		},
	}
}

func (c *Client) Resolve(domain string) ([]string, error) {
	res, err := c.QueryMultiple(domain, []uint16{dns.TypeA, dns.TypeAAAA})
	if err != nil {
		return nil, err
	}
	r := make([]string, len(res))
	for i := range res {
		r[i] = res[i].Value
	}
	return r, nil
}

func (c *Client) QueryMultiple(host string, types []uint16) ([]DNSRecord, error) {
	var res []DNSRecord

	msg := &dns.Msg{}
	for _, t := range types {
		msg.SetQuestion(dns.CanonicalName(host), t)
		resp, err := c.do(msg)
		if err != nil || resp == nil {
			continue
		}

		for _, rr := range resp.Answer {
			res = append(res, DNSRecord{
				Domain: host,
				Type:   dns.TypeToString[rr.Header().Rrtype],
				Value:  rr.String(),
			})
		}
	}

	return res, nil
}

func (c *Client) do(msg *dns.Msg) (*dns.Msg, error) {
	var resp *dns.Msg
	var err error

	for i := 0; i < c.maxRetries; i++ {
		resolver := c.resolvers[i%len(c.resolvers)]
		switch resolver.Proto() {
		case "udp":
			resp, _, err = c.udpClient.Exchange(msg, resolver.Addr())
		case "tcp":
			resp, _, err = c.tcpClient.Exchange(msg, resolver.Addr())
		}
		if err != nil || resp == nil {
			continue
		}
		if resp.Rcode != dns.RcodeSuccess {
			continue
		}
		return resp, nil
	}

	return resp, ErrMaxRetries
}
