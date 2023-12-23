package p2pnet

import (
	"BlockChain/src/mycrypto"
	"BlockChain/src/utils"
	"context"
	"errors"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"log"
	"sync"
	"time"
)

// RecvHandler receive callback func
type RecvHandler func(t MessageType, msgBytes []byte, peerID string)

type P2PNet struct {
	Host       host.Host // local node host
	ID         string
	protocol   string // p2p network protocol
	rendezvous string // rendezvous string
	//bootstrapPeers   []multiaddr.Multiaddr // bootstrap peers address
	//bootstrap        bool                  // is bootstrap peer
	//kademliaDHT      *dht.IpfsDHT          // KDH table
	//routingDiscovery *drouting.RoutingDiscovery
	sync.RWMutex                             // lock
	peerTable    map[string]*P2PStream       // already connect peers
	callBacks    map[MessageType]RecvHandler // receive call back func
	log          *log.Logger
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
func CreateNode(keyPath string, addr string, logPath string) *P2PNet {
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

	host, err := libp2p.New(
		libp2p.Identity(pri),
		libp2p.ListenAddrStrings(addr),
	)
	if err != nil {
		l.Panic(err)
	}

	p2pNode := &P2PNet{
		Host:       host,
		ID:         host.ID().String(),
		protocol:   "/chain/1.0.0",
		rendezvous: "chain-discovery",
		peerTable:  make(map[string]*P2PStream),
		callBacks:  make(map[MessageType]RecvHandler),
		log:        l,
	}
	return p2pNode
}

// StartNode start a p2p node and discovery other node
func (node *P2PNet) StartNode() {

	node.log.Println("Start p2p node")
	// register stream handler func
	node.Host.SetStreamHandler(protocol.ID(node.protocol), node.handleStream)

	ctx := context.Background()
	node.log.Println("Initialize mDNS")
	peerChan := initMDNS(node.Host, node.rendezvous)
	for {
		info := <-peerChan
		node.log.Println("Found peer:", info, ", connecting")

		if err := node.Host.Connect(ctx, info); err != nil {
			node.log.Println("Connection failed:", err)
			continue
		}

		stream, err := node.Host.NewStream(context.Background(), info.ID, protocol.ID(node.protocol))
		if err != nil {
			node.log.Println("Connection failed:", err)
			continue
		} else {
			node.log.Println("Connected to peer:", info)
			p2pStaeam := &P2PStream{
				peerID:          info.ID.String(),
				stream:          stream,
				MessageChan:     make(chan *Message),
				closeReadStrem:  make(chan struct{}, 1),
				closeWriteStrem: make(chan struct{}, 1),
			}
			node.Lock()
			node.log.Printf("add peer %s", info.ID.String())
			node.peerTable[info.ID.String()] = p2pStaeam
			node.Unlock()
			go node.recvData(p2pStaeam)
			go node.sendData(p2pStaeam)
		}
	}
}

// Broadcast message to all peers
func (node *P2PNet) Broadcast(msg *Message) error {
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
	stream, ok := node.peerTable[peerID]
	if !ok {
		return errors.New("Peer not found")
	}
	go func() {
		select {
		// send message
		case stream.MessageChan <- msg:
			node.log.Printf("send message to peer: %s, type: %v", peerID, msg.Type)
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
	node.log.Printf("Register Callback func type: %v", t)
	node.Lock()
	defer node.Unlock()
	node.callBacks[t] = callback
}
