package arp

type Arp interface {
	Close()

	GetSpoofedIP() string
	GetSpoofedMAC() string
	GetTargetMAC() string

	SetParameter(targetMACString string,
		spoofedIPString string,
		spoofedMACString string) error
	Do() error
}
