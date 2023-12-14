package state

import (
	"badger"
	"sync"
)

type WorldState struct {
	// block chain
	BlockHeight uint64
	Tip         []byte
	// pBFT
	IsPrimary    bool
	PrimaryID    string
	SelfID       string
	View         uint64
	CheckPoint   uint64
	WaterHead    uint64
	MaxFaultNode int
	db           *badger.DB
	lock         sync.Mutex
}

//
//func NewWorldState(path string) *WorldState {
//	// load world state from database
//	//db := blockchain.OpenDatabase(path)
//
//	return nil
//}
