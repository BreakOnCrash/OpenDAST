package subprober

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/BreakOnCrash/opendast/dns/client"
)

const DefaultPool = 10

type Config struct {
	Dict string `json:"dict" yaml:"dict"` // 字典文件路径
	Pool int    `json:"pool" yaml:"pool"` // 处理池子数量
}

type Prober struct {
	cfg  *Config
	dnsc *client.Client
}

func New(cfg *Config, client *client.Client) *Prober {
	if cfg.Pool <= 0 {
		cfg.Pool = DefaultPool
	}

	return &Prober{
		cfg:  cfg,
		dnsc: client,
	}
}

func (p *Prober) Probe(ctx context.Context, domain string) ([]string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	subs, err := p.productSubs(ctx)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	out := make(chan string)

	wg.Add(p.cfg.Pool)
	for i := 0; i < p.cfg.Pool; i++ {
		go func(ctx context.Context) {
			defer wg.Done()

			for {
				select {
				case sub, ok := <-subs:
					if sub == "" && !ok {
						return
					}
					sub = fmt.Sprintf("%s.%s", sub, domain)
					res, err := p.dnsc.Resolve(sub)
					if err != nil {
						continue
					}
					if len(res) > 0 {
						out <- sub
					}
				case <-ctx.Done():
					return
				}
			}
		}(ctx)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	res := make([]string, 0)
	for r := range out {
		res = append(res, r)
	}

	return res, nil
}

func (p *Prober) productSubs(ctx context.Context) (<-chan string, error) {
	// TODO
	// 目前将字典内容一次性读取出来
	data, err := os.ReadFile(p.cfg.Dict)
	if err != nil {
		return nil, err
	}

	subs := make(chan string)
	go func() {
		defer close(subs)

		s := bufio.NewScanner(bytes.NewBuffer(data))
		for s.Scan() {
			select {
			case subs <- strings.TrimSpace(s.Text()):
			case <-ctx.Done():
				return
			}
		}
	}()

	return subs, nil
}
