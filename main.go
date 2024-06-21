package main

import (
	"custom/arp"
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	defaultTargetMACString  = "00:15:5D:09:B8:34"
	defaultSpoofedIpString  = "200.201.202.144"
	defaultSpoofedMACString = "DE:AD:BE:EF:11:12"
	defaultInterfaceString  = "eth1"
)

func main() {
	targetMACString := flag.String("targetMAC", defaultTargetMACString, "The targeted machine on network identified by its MAC address")
	spoofMACString := flag.String("spoofedMAC", defaultSpoofedMACString, "The spoofed MAC address")
	spoofedIPString := flag.String("spoofedIP", defaultSpoofedIpString, "The spoofed IP address")
	interfaceString := flag.String("interface", defaultInterfaceString, "Name of the interface to use. To list interfaces, use -listInterfaces")

	listInterfaces := flag.Bool("listInterfaces", false, "List all interfaces")

	help := flag.Bool("help", false, "Show help")

	if flag.Parse(); !flag.Parsed() {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *help {
		fmt.Println("On windows, must be run as administrator")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *listInterfaces {
		arp.ListInterfaces()
		os.Exit(0)
	}

	if err := arp.Do(*interfaceString, *targetMACString, *spoofedIPString, *spoofMACString); err != nil {
		log.Fatal("Failed with error :", err)
	}
}
