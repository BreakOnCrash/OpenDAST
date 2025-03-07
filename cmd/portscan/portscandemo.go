package main

import (
	"log"

	"github.com/BreakOnCrash/opendast/portscan"
)

func main() {
	targetIP := "TODO"
	if err := portscan.TCPSYNScanDemo(targetIP, 80); err != nil {
		log.Fatal(err)
	}
}
