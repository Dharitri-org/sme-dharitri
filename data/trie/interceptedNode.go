package trie

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	logger "github.com/Dharitri-org/sme-logger"
)

var _ process.TxValidatorHandler = (*InterceptedTrieNode)(nil)
var _ process.InterceptedData = (*InterceptedTrieNode)(nil)

// InterceptedTrieNode implements intercepted data interface and is used when trie nodes are intercepted
type InterceptedTrieNode struct {
	node    node
	encNode []byte
	hash    []byte
	mutex   sync.Mutex
}

// NewInterceptedTrieNode creates a new instance of InterceptedTrieNode
func NewInterceptedTrieNode(
	buff []byte,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (*InterceptedTrieNode, error) {
	if len(buff) == 0 {
		return nil, ErrValueTooShort
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}

	n, err := decodeNode(buff, marshalizer, hasher)
	if err != nil {
		return nil, err
	}
	n.setDirty(true)

	err = n.setHash()
	if err != nil {
		return nil, err
	}

	return &InterceptedTrieNode{
		node:    n,
		encNode: buff,
		hash:    n.getHash(),
	}, nil
}

// CheckValidity checks if the intercepted data is valid
func (inTn *InterceptedTrieNode) CheckValidity() error {
	if inTn.node.isValid() {
		return nil
	}
	return ErrInvalidNode
}

// IsForCurrentShard checks if the intercepted data is for the current shard
func (inTn *InterceptedTrieNode) IsForCurrentShard() bool {
	return true
}

// Hash returns the hash of the intercepted node
func (inTn *InterceptedTrieNode) Hash() []byte {
	inTn.mutex.Lock()
	defer inTn.mutex.Unlock()

	return inTn.hash
}

// IsInterfaceNil returns true if there is no value under the interface
func (inTn *InterceptedTrieNode) IsInterfaceNil() bool {
	return inTn == nil
}

// EncodedNode returns the intercepted encoded node
func (inTn *InterceptedTrieNode) EncodedNode() []byte {
	return inTn.encNode
}

// Type returns the type of this intercepted data
func (inTn *InterceptedTrieNode) Type() string {
	return "intercepted trie node"
}

// String returns the trie node's most important fields as string
func (inTn *InterceptedTrieNode) String() string {
	return fmt.Sprintf("hash=%s",
		logger.DisplayByteSlice(inTn.hash),
	)
}

// SenderShardId returns 0
func (inTn *InterceptedTrieNode) SenderShardId() uint32 {
	return 0
}

// ReceiverShardId returns 0
func (inTn *InterceptedTrieNode) ReceiverShardId() uint32 {
	return 0
}

// Nonce return 0
func (inTn *InterceptedTrieNode) Nonce() uint64 {
	return 0
}

// SenderAddress returns nil
func (inTn *InterceptedTrieNode) SenderAddress() []byte {
	return nil
}

// Fee returns big.NewInt(0)
func (inTn *InterceptedTrieNode) Fee() *big.Int {
	return big.NewInt(0)
}

// Identifiers returns the identifiers used in requests
func (inTn *InterceptedTrieNode) Identifiers() [][]byte {
	return [][]byte{inTn.hash}
}
