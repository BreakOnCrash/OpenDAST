//go:build darwin

package device

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type Device struct {
	Name       string
	IPv4       net.IP
	MAC        net.HardwareAddr
	GatewayMAC net.HardwareAddr
}

func FindLocalNetDevice() (dev Device, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return dev, err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 ||
			strings.HasPrefix(iface.Name, "lo") ||
			strings.HasPrefix(iface.Name, "docker") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return dev, err
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				dev.IPv4 = ipnet.IP
				dev.Name = iface.Name

				dev.GatewayMAC = iface.HardwareAddr
				addr, err := getTrueMAC(dev.Name)
				if err != nil {
					return dev, err
				}
				if dev.MAC, err = net.ParseMAC(addr); err != nil {
					return dev, err
				}
				return dev, nil
			}
		}
	}

	return dev, errors.New("not found host IP")
}

func getTrueMAC(iface string) (string, error) {
	output, err := exec.Command("networksetup", "-getmacaddress", iface).Output()
	if err != nil {
		return "", err
	}
	fields := strings.Fields(strings.TrimSpace(string(output)))
	for _, field := range fields {
		// MAC 地址格式为 xx:xx:xx:xx:xx:xx，长度为 17
		if len(field) == 17 && strings.Count(field, ":") == 5 {
			return field, nil
		}
	}

	return "", fmt.Errorf("MAC address not found")
}
