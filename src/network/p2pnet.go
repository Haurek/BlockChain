package p2pnet

import (
	"BlockChain/src/mycrypto"
	"context"
	"errors"
	"github.com/ipfs/go-log/v2"
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
	"sync"
	"time"
)

var logger = log.Logger("Block Chain")

type addrList []multiaddr.Multiaddr

// RecvHandler receive callback func
type RecvHandler func(t MessageType, msgBytes []byte, peerID string)

type P2PNet struct {
	Host             host.Host    // local node host
	protocol         string       // p2p network protocol
	rendezvous       string       // rendezvous string
	bootstrapPeers   addrList     // bootstrap peers address
	bootstrap        bool         // is bootstrap peer
	kademliaDHT      *dht.IpfsDHT // KDH table
	routingDiscovery *drouting.RoutingDiscovery
	sync.RWMutex                                 // lock
	peerTable        map[string]*P2PStream       // already connect peers
	callBacks        map[MessageType]RecvHandler // receive call back func
	//specifyNodes []string
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
func CreateNode(cfg Config) (*P2PNet, error) {
	// load local private key
	privateKey, err := mycrypto.LoadPrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, err
	}
	pri, _, err := crypto.ECDSAKeyPairFromKey(privateKey)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	host, err := libp2p.New(
		libp2p.Identity(pri),
		libp2p.ListenAddrs([]multiaddr.Multiaddr(cfg.ListenAddresses)...),
	)
	if err != nil {
		return nil, err
	}

	p2pNode := &P2PNet{
		Host:       host,
		protocol:   "/chain/1.0.0",
		rendezvous: "chain-discovery",
		peerTable:  make(map[string]*P2PStream),
	}

	p2pNode.bootstrap = cfg.Bootstrap
	p2pNode.bootstrapPeers = cfg.BootstrapPeers

	// Start a DHT, for use in peer discovery
	kademliaDHT, err := dht.New(ctx, p2pNode.Host)
	if err != nil {
		return nil, err
	}
	p2pNode.kademliaDHT = kademliaDHT

	// broadcast
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	p2pNode.routingDiscovery = routingDiscovery

	return p2pNode, nil
}

// StartNode start a p2p node and discovery other node
func (node *P2PNet) StartNode() error {
	node.Host.SetStreamHandler(protocol.ID(node.protocol), node.handleStream)
	ctx := context.Background()
	// 启动分布式hash表
	if err := node.kademliaDHT.Bootstrap(ctx); err != nil {
		return err
	}
	// current node had bootstrap
	if node.bootstrap {
		return nil
	}

	// connect to bootstrap node
	var wg sync.WaitGroup
	for _, peerAddr := range node.bootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := node.Host.Connect(ctx, *peerinfo); err != nil {
				logger.Warning(err)
			} else {
				logger.Info("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()
	dutil.Advertise(ctx, node.routingDiscovery, node.rendezvous)

	// discovery node
	go node.nodeDiscovery()

	return nil
}

func (node *P2PNet) nodeDiscovery() {
	peerChan, err := node.routingDiscovery.FindPeers(context.Background(), node.rendezvous)
	if err != nil {
		panic(err)
	}
	for peer := range peerChan {
		if peer.ID == node.Host.ID() {
			continue
		}
		logger.Debug("Found peer:", peer)
		logger.Debug("Connecting to:", peer)
		if _, ok := node.peerTable[peer.ID.String()]; ok {
			logger.Debug("peer %s already connect:", peer.ID.String())
			continue
		}

		stream, err := node.Host.NewStream(context.Background(), peer.ID, protocol.ID(node.protocol))
		if err != nil {
			logger.Warning("Connection failed:", err)
			continue
		} else {
			p2pStaeam := &P2PStream{
				peerID:          peer.ID.String(),
				stream:          stream,
				MessageChan:     make(chan *Message, 0),
				closeReadStrem:  make(chan struct{}, 1),
				closeWriteStrem: make(chan struct{}, 1),
			}
			node.Lock()
			node.peerTable[peer.ID.String()] = p2pStaeam
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
			logger.Debug(err.Error())
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
			logger.Debug("send message to peer: %s", peerID)
		case <-time.After(1 * time.Minute):
			logger.Debug("Timeout!: %s", peerID)
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
			logger.Debug(err.Error())
		}
	}
	return nil
}

// RegisterCallback register callback func
func (node *P2PNet) RegisterCallback(t MessageType, callback RecvHandler) {
	node.Lock()
	node.callBacks[t] = callback
	node.Unlock()
}
