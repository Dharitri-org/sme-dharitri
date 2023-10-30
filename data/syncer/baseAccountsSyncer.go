package syncer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/trie"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/storage"
)

type baseAccountsSyncer struct {
	hasher               hashing.Hasher
	marshalizer          marshal.Marshalizer
	trieSyncers          map[string]data.TrieSyncer
	dataTries            map[string]data.Trie
	mutex                sync.Mutex
	trieStorageManager   data.StorageManager
	requestHandler       trie.RequestHandler
	waitTime             time.Duration
	shardId              uint32
	cacher               storage.Cacher
	rootHash             []byte
	maxTrieLevelInMemory uint
}

const minWaitTime = time.Second

// ArgsNewBaseAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewBaseAccountsSyncer struct {
	Hasher               hashing.Hasher
	Marshalizer          marshal.Marshalizer
	TrieStorageManager   data.StorageManager
	RequestHandler       trie.RequestHandler
	WaitTime             time.Duration
	Cacher               storage.Cacher
	MaxTrieLevelInMemory uint
}

func checkArgs(args ArgsNewBaseAccountsSyncer) error {
	if check.IfNil(args.Hasher) {
		return state.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return state.ErrNilMarshalizer
	}
	if check.IfNil(args.TrieStorageManager) {
		return state.ErrNilStorageManager
	}
	if check.IfNil(args.RequestHandler) {
		return state.ErrNilRequestHandler
	}
	if args.WaitTime < minWaitTime {
		return fmt.Errorf("%w, minWaitTime is %d", state.ErrInvalidWaitTime, minWaitTime)
	}
	if check.IfNil(args.Cacher) {
		return state.ErrNilCacher
	}

	return nil
}

func (b *baseAccountsSyncer) syncMainTrie(rootHash []byte, trieTopic string, ctx context.Context) error {
	b.rootHash = rootHash

	dataTrie, err := trie.NewTrie(b.trieStorageManager, b.marshalizer, b.hasher, b.maxTrieLevelInMemory)
	if err != nil {
		return err
	}

	b.dataTries[string(rootHash)] = dataTrie
	trieSyncer, err := trie.NewTrieSyncer(b.requestHandler, b.cacher, dataTrie, b.shardId, trieTopic)
	if err != nil {
		return err
	}
	b.trieSyncers[string(rootHash)] = trieSyncer

	err = trieSyncer.StartSyncing(rootHash, ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetSyncedTries returns the synced map of data trie
func (b *baseAccountsSyncer) GetSyncedTries() map[string]data.Trie {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	clonedMap := make(map[string]data.Trie, len(b.dataTries))
	for key, value := range b.dataTries {
		clonedMap[key] = value
	}

	return clonedMap
}

// IsInterfaceNil returns true if underlying object is nil
func (b *baseAccountsSyncer) IsInterfaceNil() bool {
	return b == nil
}
