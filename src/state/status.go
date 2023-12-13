package state

type WorldState struct {
	// block chain
	BlockHeight uint64
	PreHash     []byte
	// pBFT
	IsPrimary        bool
	PrimaryID        string
	View             uint64
	CurrentView      uint64
	StableCheckPoint uint64
	CheckPoint       uint64
	WaterHead        uint64
	MaxFaultNode     int
}

func NewWorldState() *WorldState {
	// load world state from database
	// TODO
	return nil
}
