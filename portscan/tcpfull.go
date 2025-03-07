package portscan

import (
	"fmt"
	"net"
	"time"
)

func TCPFullScan(ip net.IP, port uint16) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip.String(), port), 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}
