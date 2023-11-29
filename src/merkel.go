package BlockChain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// MerkelNode merkel tree node
type MerkelNode struct {
	Left  *MerkelNode
	Right *MerkelNode
	Hash  []byte
}

// MerkelTree merkel tree
type MerkelTree struct {
	Root *MerkelNode
}

func (tree *MerkelTree) NewMerkelTree(datas [][]byte) *MerkelTree {
	var nodes []*MerkelNode

	// Create leaf nodes
	for _, data := range datas {
		node := &MerkelNode{}
		node.NewMerkelNode(nil, nil, data)
		nodes = append(nodes, node)
	}

	// Build the tree
	for len(nodes) > 1 {
		var newLevel []*MerkelNode

		// Handle odd number of nodes
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		// Combine pairs and create parent nodes
		for i := 0; i < len(nodes); i += 2 {
			node := &MerkelNode{}
			node.NewMerkelNode(nodes[i], nodes[i+1], nil)
			newLevel = append(newLevel, node)
		}

		nodes = newLevel
	}

	tree.Root = nodes[0]

	return tree
}

func (node *MerkelNode) NewMerkelNode(left, right *MerkelNode, data []byte) *MerkelNode {
	newNode := &MerkelNode{Left: left, Right: right}

	if data != nil {
		hash := sha256.Sum256(data)
		newNode.Hash = hash[:]
	} else {
		// If data is nil, it's an internal node
		hash := sha256.Sum256(append(left.Hash, right.Hash...))
		newNode.Hash = hash[:]
	}

	return newNode
}

func main() {
	// Example usage
	data := [][]byte{
		[]byte("Transaction1"),
		[]byte("Transaction2"),
		[]byte("Transaction3"),
		[]byte("Transaction4"),
	}

	merkelTree := &MerkelTree{}
	merkelTree.NewMerkelTree(data)

	// Print the root hash of the Merkel Tree
	fmt.Println("Merkel Root Hash:", hex.EncodeToString(merkelTree.Root.Hash))
}
