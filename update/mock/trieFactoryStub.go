package mock

import (
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/data"
)

// TrieFactoryStub -
type TrieFactoryStub struct {
	CreateCalled func(config config.StorageConfig, s string, b bool) (data.StorageManager, data.Trie, error)
}

// Create -
func (t *TrieFactoryStub) Create(config config.StorageConfig, s string, b bool) (data.StorageManager, data.Trie, error) {
	if t.CreateCalled != nil {
		return t.CreateCalled(config, s, b)
	}
	return nil, nil, nil
}

// IsInterfaceNil -
func (t *TrieFactoryStub) IsInterfaceNil() bool {
	return t == nil
}
