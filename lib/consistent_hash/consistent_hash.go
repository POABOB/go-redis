package consistent_hash

import (
	"hash/crc32"
	"sort"
)

// NodeFunc is a function that returns the hash value of a given data
type NodeFunc func(data []byte) uint32

// NodeMap is a map of nodes
type NodeMap struct {
	hashFunc   NodeFunc
	NodeHashes []int // slice will be sorted
	Nodes      map[int]string
}

// NewNodeMap returns a new NodeMap
func NewNodeMap(hashFunc NodeFunc) *NodeMap {
	nodeMap := &NodeMap{
		hashFunc: hashFunc,
		Nodes:    make(map[int]string),
	}
	if nodeMap.hashFunc == nil {
		nodeMap.hashFunc = crc32.ChecksumIEEE
	}
	return nodeMap
}

// AddNode adds nodes to the NodeMap and sorts them
func (nodeMap *NodeMap) AddNode(nodes ...string) {
	for _, node := range nodes {
		if node == "" {
			continue
		}
		hash := int(nodeMap.hashFunc([]byte(node)))
		nodeMap.Nodes[hash] = node
		nodeMap.NodeHashes = append(nodeMap.NodeHashes, hash)
	}
	sort.Ints(nodeMap.NodeHashes)
}

// IsEmpty returns true if the NodeMap is empty
func (nodeMap *NodeMap) IsEmpty() bool {
	return len(nodeMap.NodeHashes) == 0
}

// GetNode returns the node with the given key value
func (nodeMap *NodeMap) GetNode(key string) string {
	if nodeMap.IsEmpty() {
		return ""
	}
	hash := int(nodeMap.hashFunc([]byte(key)))
	index := sort.Search(len(nodeMap.NodeHashes), func(i int) bool {
		return nodeMap.NodeHashes[i] >= hash
	})
	index = index % len(nodeMap.NodeHashes)
	return nodeMap.Nodes[nodeMap.NodeHashes[index]]
}
