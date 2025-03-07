package portscan

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/BreakOnCrash/opendast/portscan/device"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func TCPSYNScanDemo(targetIP string, targetPort uint16) error {
	localdev, err := device.FindLocalNetDevice()
	if err != nil {
		return err
	}

	srcPort, err := GetFreePort()
	if err != nil {
		return err
	}

	handle, err := pcap.OpenLive(localdev.Name, 1600, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	// send SYN packet
	ethLayer := &layers.Ethernet{
		SrcMAC:       localdev.MAC,
		DstMAC:       localdev.GatewayMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}

	ipLayer := &layers.IPv4{
		SrcIP:    localdev.IPv4,
		DstIP:    net.ParseIP(targetIP),
		Protocol: layers.IPProtocolTCP,
		Version:  4,
		TTL:      64,
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tcpLayer := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(targetPort),
		SYN:     true,
		Seq:     r.Uint32(),
		Window:  14600,
	}

	tcpLayer.SetNetworkLayerForChecksum(ipLayer)

	buffer := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buffer, opts, ethLayer, ipLayer, tcpLayer); err != nil {
		return err
	}
	if err := handle.WritePacketData(buffer.Bytes()); err != nil {
		return err
	}

	// receive packet
	timeout := time.After(5 * time.Second)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for {
		select {
		case packet := <-packetSource.Packets():
			tcpLayer := packet.Layer(layers.LayerTypeTCP)
			if tcpLayer == nil {
				continue
			}
			tcp, _ := tcpLayer.(*layers.TCP)

			ipLayer := packet.Layer(layers.LayerTypeIPv4)
			if ipLayer == nil {
				continue
			}
			ip, _ := ipLayer.(*layers.IPv4)

			if ip.SrcIP.String() == targetIP && ip.DstIP.String() == localdev.IPv4.String() &&
				uint16(tcp.SrcPort) == targetPort && uint16(tcp.DstPort) == srcPort {

				if tcp.SYN && tcp.ACK {
					fmt.Printf("Port %d is open (SYN-ACK received)\n", targetPort)
				} else if tcp.RST {
					fmt.Printf("Port %d is closed (RST received)\n", targetPort)
				}
				return nil
			}
		case <-timeout:
			fmt.Printf("Port %d filtered or no response (timeout)\n", targetPort)
			return nil
		}
	}
}
