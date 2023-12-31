package state

import (
	"github.com/Dharitri-org/sme-dharitri/data"
)

type baseAccount struct {
	address         []byte
	code            []byte
	dataTrieTracker DataTrieTracker
}

// AddressBytes returns the address associated with the account as byte slice
func (ba *baseAccount) AddressBytes() []byte {
	return ba.address
}

// GetCode gets the actual code that needs to be run in the VM
func (ba *baseAccount) GetCode() []byte {
	return ba.code
}

// SetCode sets the actual code that needs to be run in the VM
func (ba *baseAccount) SetCode(code []byte) {
	ba.code = code
}

// DataTrie returns the trie that holds the current account's data
func (ba *baseAccount) DataTrie() data.Trie {
	return ba.dataTrieTracker.DataTrie()
}

// SetDataTrie sets the trie that holds the current account's data
func (ba *baseAccount) SetDataTrie(trie data.Trie) {
	ba.dataTrieTracker.SetDataTrie(trie)
}

// DataTrieTracker returns the trie wrapper used in managing the SC data
func (ba *baseAccount) DataTrieTracker() DataTrieTracker {
	return ba.dataTrieTracker
}

// IsInterfaceNil returns true if there is no value under the interface
func (ba *baseAccount) IsInterfaceNil() bool {
	return ba == nil
}
