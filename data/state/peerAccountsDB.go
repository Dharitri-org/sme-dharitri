package state

import (
	"sync"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
)

// PeerAccountsDB will save and synchronize data from peer processor, plus will synchronize with nodesCoordinator
type PeerAccountsDB struct {
	*AccountsDB
}

// NewPeerAccountsDB creates a new account manager
func NewPeerAccountsDB(
	trie data.Trie,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	accountFactory AccountFactory,
) (*PeerAccountsDB, error) {
	if check.IfNil(trie) {
		return nil, ErrNilTrie
	}
	if check.IfNil(hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(accountFactory) {
		return nil, ErrNilAccountFactory
	}

	numCheckpoints := getNumCheckpoints(trie)
	return &PeerAccountsDB{
		&AccountsDB{
			mainTrie:       trie,
			hasher:         hasher,
			marshalizer:    marshalizer,
			accountFactory: accountFactory,
			entries:        make([]JournalEntry, 0),
			dataTries:      NewDataTriesHolder(),
			mutOp:          sync.RWMutex{},
			numCheckpoints: numCheckpoints,
		},
	}, nil
}

// SnapshotState triggers the snapshotting process of the state trie
func (adb *PeerAccountsDB) SnapshotState(rootHash []byte) {
	log.Trace("peerAccountsDB.SnapshotState", "root hash", rootHash)
	adb.mainTrie.EnterSnapshotMode()
	adb.mainTrie.TakeSnapshot(rootHash)
	adb.mainTrie.ExitSnapshotMode()

	adb.increaseNumCheckpoints()
}

// SetStateCheckpoint triggers the checkpointing process of the state trie
func (adb *PeerAccountsDB) SetStateCheckpoint(rootHash []byte) {
	log.Trace("peerAccountsDB.SetStateCheckpoint", "root hash", rootHash)
	adb.mainTrie.EnterSnapshotMode()
	adb.mainTrie.SetCheckpoint(rootHash)
	adb.mainTrie.ExitSnapshotMode()

	adb.increaseNumCheckpoints()
}

// RecreateAllTries recreates all the tries from the accounts DB
func (adb *PeerAccountsDB) RecreateAllTries(rootHash []byte) (map[string]data.Trie, error) {
	recreatedTrie, err := adb.mainTrie.Recreate(rootHash)
	if err != nil {
		return nil, err
	}

	allTries := make(map[string]data.Trie)
	allTries[string(rootHash)] = recreatedTrie

	return allTries, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (adb *PeerAccountsDB) IsInterfaceNil() bool {
	return adb == nil
}
