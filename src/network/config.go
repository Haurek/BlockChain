package p2pnet

type Config struct {
	// P2P
	RendezvousString string
	BootstrapPeers   addrList
	ListenAddresses  addrList
	ProtocolID       string
	PrivateKey       string
	Bootstrap        bool
}
