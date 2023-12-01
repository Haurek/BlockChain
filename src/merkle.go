package BlockChain

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
func NewMerkleTree(datas [][]byte) *MerkleTree {
	var nodes []MerkleNode

	// Create leaf nodes
	for _, data := range datas {
		node := NewMerkleNode(nil, nil, data)
		nodes = append(nodes, *node)
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
			node := NewMerkleNode(&nodes[i], &nodes[i+1], nil)
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	tree := &MerkleTree{
		Root: &nodes[0],
	}

	return tree
}

// NewMerkleNode create merkle tree node
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	var hash []byte
	if data != nil {
		hash = Sha256Hash(data)
	} else {
		// If data is nil, it's an internal node
		hash = Sha256Hash(append(left.Hash, right.Hash...))
	}
	node := &MerkleNode{
		Hash: hash,
	}
	return node
}
