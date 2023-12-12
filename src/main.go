package main

import (
	"flag"
	"fmt"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	maddr "github.com/multiformats/go-multiaddr"
	"strings"
)

// A new type we need for writing a custom flag parser
type addrList []maddr.Multiaddr

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

func StringsToAddrs(addrStrings []string) (maddrs []maddr.Multiaddr, err error) {
	for _, addrString := range addrStrings {
		addr, err := maddr.NewMultiaddr(addrString)
		if err != nil {
			return maddrs, err
		}
		maddrs = append(maddrs, addr)
	}
	return
}

type Config struct {
	// P2P
	RendezvousString string
	BootstrapPeers   addrList
	ListenAddresses  addrList
	ProtocolID       string
	PrivateKey       string
	Bootstrap        bool
}

func ParseFlags() (Config, error) {
	config := Config{}
	flag.StringVar(&config.RendezvousString, "rendezvous", "Block Chain",
		"Unique string to identify group of nodes.")
	flag.Var(&config.BootstrapPeers, "p", "Adds a peer multiaddress to the bootstrap list")
	flag.Var(&config.ListenAddresses, "c", "Create a new node and set listen address")
	flag.StringVar(&config.ProtocolID, "pid", "/chain/1.0.0", "Sets a protocol id for stream headers")
	flag.Parse()

	if len(config.BootstrapPeers) == 0 {
		config.BootstrapPeers = dht.DefaultBootstrapPeers
	}
	return config, nil
}

func Run() {
	//log.SetAllLoggers(log.LevelWarn)
	//log.SetLogLevel("Block Chain", "info")
	help := flag.Bool("h", false, "Display Help")
	config, err := ParseFlags()
	if err != nil {
		panic(err)
	}
	if *help {
		flag.PrintDefaults()
		return
	}
	if len(config.ListenAddresses) != 0 {
		//p2pnet.CreateNode(config)
	} else {
		fmt.Println("you need to set a listen adderss")
	}
}
