package p2pnet

import (
	"BlockChain/src/mycrypto"
	"BlockChain/src/utils"
	"context"
	"errors"
	"fmt"
	p2plog "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	"log"
	"sync"
	"time"
)

var logger = p2plog.Logger("rendezvous")

// RecvHandler receive callback func
type RecvHandler func(t MessageType, msgBytes []byte, peerID string)

type P2PNet struct {
	Host             host.Host             // local node host
	protocol         string                // p2p network protocol
	rendezvous       string                // rendezvous string
	bootstrapPeers   []multiaddr.Multiaddr // bootstrap peers address
	bootstrap        bool                  // is bootstrap peer
	kademliaDHT      *dht.IpfsDHT          // KDH table
	routingDiscovery *drouting.RoutingDiscovery
	sync.RWMutex                                 // lock
	peerTable        map[string]*P2PStream       // already connect peers
	callBacks        map[MessageType]RecvHandler // receive call back func
	log              *log.Logger
}

// P2PStream describe stream between peer
type P2PStream struct {
	peerID          string
	stream          network.Stream
	MessageChan     chan *Message
	closeReadStrem  chan struct{}
	closeWriteStrem chan struct{}
}

// CreateNode create a P2P network Node
func CreateNode(keyPath string, addr string, bootstrap bool, bootstrapPeers []string, logPath string) *P2PNet {
	// initialize logger
	l := utils.NewLogger("[net] ", logPath)

	// load local private key
	privateKey, err := mycrypto.LoadPrivateKey(keyPath)
	if err != nil {
		l.Panic("Load private key failed: ")
	}
	pri, _, err := crypto.ECDSAKeyPairFromKey(privateKey)
	if err != nil {
		l.Panic("Initialize private key fail")
	}

	ctx := context.Background()
	host, err := libp2p.New(
		libp2p.Identity(pri),
		libp2p.ListenAddrStrings(addr),
	)
	if err != nil {
		l.Panic(err)
	}

	p2pNode := &P2PNet{
		Host:       host,
		protocol:   "/chain/1.0.0",
		rendezvous: "chain-discovery",
		peerTable:  make(map[string]*P2PStream),
		callBacks:  make(map[MessageType]RecvHandler),
		log:        l,
	}

	p2pNode.bootstrap = bootstrap
	var bootstraps []multiaddr.Multiaddr
	if len(bootstrapPeers) == 0 {
		bootstraps = dht.DefaultBootstrapPeers
	} else {
		for _, peerAddr := range bootstrapPeers {
			mAddr, err := multiaddr.NewMultiaddr(peerAddr)
			if err != nil {
				return nil
			}
			bootstraps = append(bootstraps, mAddr)
		}
	}

	// Start a DHT, for use in peer discovery
	kademliaDHT, err := dht.New(ctx, p2pNode.Host)
	if err != nil {
		l.Panic(err)
	}
	p2pNode.kademliaDHT = kademliaDHT

	p2pNode.routingDiscovery = drouting.NewRoutingDiscovery(kademliaDHT)
	return p2pNode
}

// StartNode start a p2p node and discovery other node
func (node *P2PNet) StartNode(done chan struct{}) {
	node.log.Println("Start p2p node")
	// register stream handler func
	node.Host.SetStreamHandler(protocol.ID(node.protocol), node.handleStream)
	ctx := context.Background()

	node.log.Println("start DHT")
	// start DHT
	if err := node.kademliaDHT.Bootstrap(ctx); err != nil {
		node.log.Panic(err)
	}

	// current node had bootstrap
	if node.bootstrap {
		node.log.Println("P2P net had started")
	}
	node.bootstrap = true

	// connect to bootstrap node
	for _, peerAddr := range node.bootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		func() {
			if err := node.Host.Connect(ctx, *peerinfo); err != nil {
				node.log.Println(err)
				//logger.Warning(err)
			} else {
				node.log.Println("Connection established with bootstrap node:", *peerinfo)
				logger.Info("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	node.log.Println("Broadcast")
	dutil.Advertise(ctx, node.routingDiscovery, node.rendezvous)

	go node.nodeDiscovery()

	close(done)
}

func (node *P2PNet) nodeDiscovery() {
	// discovery node
	node.log.Println("Discovery")
	peerChan, err := node.routingDiscovery.FindPeers(context.Background(), node.rendezvous)
	if err != nil {
		node.log.Panic(err)
	}
	for info := range peerChan {
		if info.ID == node.Host.ID() {
			continue
		}
		node.log.Println("Found info: ", info)
		node.log.Println("Connecting to: ", info)
		//logger.Debug("Found peer:", info)
		//logger.Debug("Connecting to:", info)
		fmt.Println("Connecting to: ", info)
		if _, ok := node.peerTable[info.ID.String()]; ok {
			node.log.Println("info %s already connect:", info.ID.String())
			continue
		}

		stream, err := node.Host.NewStream(context.Background(), info.ID, protocol.ID(node.protocol))
		if err != nil {
			node.log.Println("Connection failed:", err)
			continue
		} else {
			p2pStaeam := &P2PStream{
				peerID:          info.ID.String(),
				stream:          stream,
				MessageChan:     make(chan *Message, 0),
				closeReadStrem:  make(chan struct{}, 1),
				closeWriteStrem: make(chan struct{}, 1),
			}
			node.Lock()
			node.peerTable[info.ID.String()] = p2pStaeam
			node.Unlock()
			go node.recvData(p2pStaeam)
			go node.sendData(p2pStaeam)
			//rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
			//
			//go node.writeData(rw)
			//go node.readData(rw)
		}
	}
	select {}
}

// Broadcast message to all peers
func (node *P2PNet) Broadcast(msg *Message) error {
	node.log.Println("Broadcast message, type: ", msg.Type)
	for id := range node.peerTable {
		err := node.BroadcastToPeer(msg, id)
		if err != nil {
			node.log.Println(err.Error())
		}
	}
	return nil
}

// BroadcastToPeer broadcast message to a peer
func (node *P2PNet) BroadcastToPeer(msg *Message, peerID string) error {
	node.log.Println("Broadcast message to peer %s", peerID)
	stream, ok := node.peerTable[peerID]
	if !ok {
		return errors.New("Peer not found")
	}
	go func() {
		select {
		// send message
		case stream.MessageChan <- msg:
			node.log.Println("send message to peer: %s", peerID)
		case <-time.After(1 * time.Minute):
			node.log.Println("Timeout!: %s", peerID)
			return
		}
	}()
	return nil
}

// BroadcastExceptPeer broadcast message to all peers except p
func (node *P2PNet) BroadcastExceptPeer(msg *Message, peerID string) error {
	for id := range node.peerTable {
		if id == id {
			continue
		}
		err := node.BroadcastToPeer(msg, peerID)
		if err != nil {
			node.log.Println(err.Error())
		}
	}
	return nil
}

// RegisterCallback register callback func
func (node *P2PNet) RegisterCallback(t MessageType, callback RecvHandler) {
	node.Lock()
	defer node.Unlock()
	node.callBacks[t] = callback
}
