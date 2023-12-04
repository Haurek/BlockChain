package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	"sync"
	"time"
)

// var logger = log.Logger("rendezvous")
var ctx = context.Background()

type Node struct {
	Host *host.Host
}

// CreateNode initialize local chain,create node and synchronize
func CreateNode(config Config) (*Node, error) {
	// initialize block chain
	chain := LoadChain()

	// constructs a new Host.
	h, err := libp2p.New(
		libp2p.ListenAddrs([]multiaddr.Multiaddr(config.ListenAddresses)...),
	)
	if err != nil {
		panic(err)
	}

	node := &Node{
		Host: &h,
	}
	// Start a DHT, for use in peer discovery
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// connect to the bootstrap nodes first
	var wg sync.WaitGroup
	for _, peerAddr := range config.BootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	// broadcast
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, config.RendezvousString)

	// find peer
	peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)
	if err != nil {
		panic(err)
	}

	// set stream handler
	h.SetStreamHandler(protocol.ID(config.ProtocolID), func(stream network.Stream) {
		handleStream(stream, chain)
	})

	for peer := range peerChan {
		if peer.ID == h.ID() {
			continue
		}

		// find and create a connection to peer
		stream, err := h.NewStream(ctx, peer.ID, protocol.ID(config.ProtocolID))

		if err != nil {
			fmt.Println("fail to connect peer: ", peer.ID)
			continue
		} else {
			// initialize message type for gob
			InitMsg()
			// block synchronization
			go Synchronize(stream, chain)
		}
	}

	select {}
	return node, nil
}

// Synchronize block chain between peer
func Synchronize(stream network.Stream, chain *Chain) {
	// send version message for handshake and start synchronize
	version := VersionMsg{
		Version:    VersionInfo,
		BestHeight: chain.BestHeight,
		TimeStamp:  time.Now().Unix(),
	}
	payload, err := WrapMsg(Version, version)
	HandleError(err)
	SendMessage(stream, payload)
}

// SendMessage send data to peer
func SendMessage(stream network.Stream, data []byte) error {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	_, err := rw.Write(data)
	if err != nil {
		fmt.Println("Error writing to buffer")
		return err
	}
	err = rw.Flush()
	if err != nil {
		fmt.Println("Error flushing buffer")
		return err
	}
	return nil
}
