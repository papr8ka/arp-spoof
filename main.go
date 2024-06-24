package main

import (
	"flag"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/papr8ka/arp-spoof/arp"
	"github.com/papr8ka/arp-spoof/interactive"
	"github.com/papr8ka/arp-spoof/logger"
	"github.com/papr8ka/arp-spoof/logger/label"
	"github.com/papr8ka/arp-spoof/token"
	"go.uber.org/zap"
	"os"
	"time"
)

const (
	defaultTargetMACString  = "00:15:5d:09:b8:34"
	defaultSpoofedIpString  = "200.201.202.144"
	defaultSpoofedMACString = "de:ad:be:ef:11:12"
	defaultInterfaceString  = "eth1"
)

func main() {
	_ = logger.Setup()
	defer logger.Close()

	targetMACString := flag.String("targetMAC", defaultTargetMACString, "The targeted machine on network identified by its MAC address")
	spoofMACString := flag.String("spoofedMAC", defaultSpoofedMACString, "The spoofed MAC address")
	spoofedIPString := flag.String("spoofedIP", defaultSpoofedIpString, "The spoofed IP address")
	interfaceString := flag.String("interface", defaultInterfaceString, "Name of the interface to use. To list interfaces, use -listInterfaces")

	isInteractive := flag.Bool("interactive", false, "Should be interactive")
	listInterfaces := flag.Bool("listInterfaces", false, "List all interfaces")
	help := flag.Bool("help", false, "Show help")

	if flag.Parse(); !flag.Parsed() {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if !token.IsAdmin() {
		logger.Logger.Error("/!\\ You have to run this as administrator/root /!\\")
		os.Exit(1)
	}

	if instance, err := arp.New(*interfaceString); err == nil {
		defer instance.Close()

		if *listInterfaces {
			instance.ListInterfaces()
		} else {
			if err = instance.SetParameter(*targetMACString, *spoofedIPString, *spoofMACString); err != nil {
				logger.Logger.Error("invalid parameters",
					zap.Error(err),
					zap.String(label.TargetMAC, *targetMACString),
					zap.String(label.SpoofedIP, *spoofedIPString),
					zap.String(label.SpoofedMAC, *spoofMACString))
			}

			isRunning := true

			go func() {
				if err := instance.Do(); err != nil {
					logger.Logger.Error("could not run ARP spoofing",
						zap.Error(err))
					isRunning = false
				}
			}()

			if *isInteractive {
				ebiten.SetWindowTitle("ARP SPOOF - " + *interfaceString)
				ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
				if err := ebiten.RunGame(interactive.New(instance)); err != nil {
					logger.Logger.Fatal("could not start interactive window",
						zap.Error(err))
				}
			} else {
				for isRunning {
					time.Sleep(time.Second)
				}
			}
		}
	} else {
		logger.Logger.Fatal("could not create arp instance",
			zap.Error(err))
	}
}
