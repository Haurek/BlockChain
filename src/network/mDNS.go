package p2pnet

import (
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// discoveryNotifee is a struct that implements the notification interface for discovered peers.
type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo // PeerChan is a channel to send discovered peer information.
}

// HandlePeerFound handles the event when a new peer is discovered and sends its address information through PeerChan.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

// initMDNS initializes the MDNS service for peer discovery.
func initMDNS(peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	// Create a discoveryNotifee to receive notifications about discovered peers.
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo) // Initialize a channel for discovered peer information.

	// Create an MDNS service to enable peer discovery.
	ser := mdns.NewMdnsService(peerhost, rendezvous, n)
	if err := ser.Start(); err != nil {
		panic(err) // Panic if there's an error starting the MDNS service.
	}
	return n.PeerChan // Return the channel for discovered peer information.
}
