package BlockChain

import (
	"crypto/sha256"
)

// MerkleNode merkel tree node
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Hash  []byte
}

// MerkleTree merkel tree
type MerkleTree struct {
	Root *MerkleNode
}

// NewMerkleTree create merkle tree from transactions byte data
func (tree *MerkleTree) NewMerkleTree(datas [][]byte) *MerkleTree {
	var nodes []MerkleNode

	// Create leaf nodes
	for _, data := range datas {
		node := MerkleNode{}
		node.NewMerkleNode(nil, nil, data)
		nodes = append(nodes, node)
	}

	// Build the tree
	for len(nodes) > 1 {
		var newLevel []MerkleNode

		// Handle odd number of nodes
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		// Combine pairs and create parent nodes
		for i := 0; i < len(nodes); i += 2 {
			node := MerkleNode{}
			node.NewMerkleNode(&nodes[i], &nodes[i+1], nil)
			newLevel = append(newLevel, node)
		}

		nodes = newLevel
	}

	tree.Root = &nodes[0]

	return tree
}

// NewMerkleNode create merkle tree node
func (node *MerkleNode) NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	if data != nil {
		hash := sha256.Sum256(data)
		node.Hash = hash[:]
	} else {
		// If data is nil, it's an internal node
		hash := sha256.Sum256(append(left.Hash, right.Hash...))
		node.Hash = hash[:]
	}

	return node
}
