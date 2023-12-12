package state

type WorldState struct {
	// block chain
	BlockHeight uint64
	// pBFT
	View             uint64
	CurrentView      uint64
	StableCheckPoint uint64
	CheckPoint       uint64
	WaterHead        int
}

func NewWorldState() *WorldState {
	// load world state from database
	// TODO
	return nil
}
