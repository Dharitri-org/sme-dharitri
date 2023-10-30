package mock

import "github.com/Dharitri-org/sme-dharitri/data"

// AccountsDBSyncerStub -
type AccountsDBSyncerStub struct {
	GetSyncedTriesCalled func() map[string]data.Trie
	SyncAccountsCalled   func(rootHash []byte) error
}

// GetSyncedTries -
func (a *AccountsDBSyncerStub) GetSyncedTries() map[string]data.Trie {
	if a.GetSyncedTriesCalled != nil {
		return a.GetSyncedTriesCalled()
	}
	return nil
}

// SyncAccounts -
func (a *AccountsDBSyncerStub) SyncAccounts(rootHash []byte) error {
	if a.SyncAccountsCalled != nil {
		return a.SyncAccountsCalled(rootHash)
	}
	return nil
}

// IsInterfaceNil -
func (a *AccountsDBSyncerStub) IsInterfaceNil() bool {
	return a == nil
}
