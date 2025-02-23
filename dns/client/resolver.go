package client

import "strings"

type Resolver interface {
	Addr() string
	Proto() string
}

var DefaultResolvers = []string{
	"udp:1.1.1.1:53", // Cloudflare
	"udp:1.0.0.1:53", // Cloudflare
	"udp:8.8.8.8:53", // Google
	"udp:8.8.4.4:53", // Google
}

type BaseResolver struct {
	addr  string
	proto string
}

func (r BaseResolver) Addr() string {
	return r.addr
}
func (r BaseResolver) Proto() string {
	return r.proto
}

func ParseResolvers(resolvers []string) []Resolver {
	var parsed []Resolver
	for _, r := range resolvers {
		proto := "udp"
		addr := r
		if len(r) >= 4 && r[3] == ':' {
			addr = r[4:]
			switch r[0:3] {
			case "udp":
			case "tcp":
				proto = "tcp"
			default:
				// unsupported protocol?
				continue
			}
		}
		if len(addr) == 0 {
			continue
		}
		if !strings.Contains(addr, ":") {
			addr = addr + ":53"
		}
		parsed = append(parsed, BaseResolver{addr: addr, proto: proto})
	}
	return parsed
}
