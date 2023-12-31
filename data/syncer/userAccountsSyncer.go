package syncer

import (
	"context"
	"sync"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/trie"
	"github.com/Dharitri-org/sme-dharitri/epochStart"
	"github.com/Dharitri-org/sme-dharitri/process/factory"
	logger "github.com/Dharitri-org/sme-logger"
)

var _ epochStart.AccountsDBSyncer = (*userAccountsSyncer)(nil)

var log = logger.GetOrCreate("syncer")

const timeBetweenRetries = 100 * time.Millisecond

type userAccountsSyncer struct {
	*baseAccountsSyncer
	throttler   data.GoRoutineThrottler
	syncerMutex sync.Mutex
}

// ArgsNewUserAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewUserAccountsSyncer struct {
	ArgsNewBaseAccountsSyncer
	ShardId   uint32
	Throttler data.GoRoutineThrottler
}

// NewUserAccountsSyncer creates a user account syncer
func NewUserAccountsSyncer(args ArgsNewUserAccountsSyncer) (*userAccountsSyncer, error) {
	err := checkArgs(args.ArgsNewBaseAccountsSyncer)
	if err != nil {
		return nil, err
	}

	if check.IfNil(args.Throttler) {
		return nil, data.ErrNilThrottler
	}

	b := &baseAccountsSyncer{
		hasher:               args.Hasher,
		marshalizer:          args.Marshalizer,
		trieSyncers:          make(map[string]data.TrieSyncer),
		dataTries:            make(map[string]data.Trie),
		trieStorageManager:   args.TrieStorageManager,
		requestHandler:       args.RequestHandler,
		waitTime:             args.WaitTime,
		shardId:              args.ShardId,
		cacher:               args.Cacher,
		rootHash:             nil,
		maxTrieLevelInMemory: args.MaxTrieLevelInMemory,
	}

	u := &userAccountsSyncer{
		baseAccountsSyncer: b,
		throttler:          args.Throttler,
	}

	return u, nil
}

// SyncAccounts will launch the syncing method to gather all the data needed for userAccounts - it is a blocking method
func (u *userAccountsSyncer) SyncAccounts(rootHash []byte) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), u.waitTime)
	defer cancel()

	err := u.syncMainTrie(rootHash, factory.AccountTrieNodesTopic, ctx)
	if err != nil {
		return err
	}

	mainTrie := u.dataTries[string(rootHash)]
	rootHashes, err := u.findAllAccountRootHashes(mainTrie)
	if err != nil {
		return err
	}

	err = u.syncAccountDataTries(rootHashes, ctx)
	if err != nil {
		return err
	}

	return nil
}

func (u *userAccountsSyncer) syncAccountDataTries(rootHashes [][]byte, ctx context.Context) error {
	var errFound error
	errMutex := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(len(rootHashes))

	for _, rootHash := range rootHashes {
		for {
			if u.throttler.CanProcess() {
				break
			}

			select {
			case <-time.After(timeBetweenRetries):
				continue
			case <-ctx.Done():
				return data.ErrTimeIsOut
			}
		}

		go func(trieRootHash []byte) {
			newErr := u.syncDataTrie(trieRootHash, ctx)
			if newErr != nil {
				errMutex.Lock()
				errFound = newErr
				errMutex.Unlock()
			}
			wg.Done()
		}(rootHash)
	}

	wg.Wait()

	errMutex.Lock()
	defer errMutex.Unlock()

	return errFound
}

func (u *userAccountsSyncer) syncDataTrie(rootHash []byte, ctx context.Context) error {
	u.throttler.StartProcessing()

	u.syncerMutex.Lock()
	if _, ok := u.dataTries[string(rootHash)]; ok {
		u.syncerMutex.Unlock()
		u.throttler.EndProcessing()
		return nil
	}

	dataTrie, err := trie.NewTrie(u.trieStorageManager, u.marshalizer, u.hasher, u.maxTrieLevelInMemory)
	if err != nil {
		u.syncerMutex.Unlock()
		return err
	}

	u.dataTries[string(rootHash)] = dataTrie
	trieSyncer, err := trie.NewTrieSyncer(u.requestHandler, u.cacher, dataTrie, u.shardId, factory.AccountTrieNodesTopic)
	if err != nil {
		u.syncerMutex.Unlock()
		return err
	}
	u.trieSyncers[string(rootHash)] = trieSyncer
	u.syncerMutex.Unlock()

	err = trieSyncer.StartSyncing(rootHash, ctx)
	if err != nil {
		return err
	}

	u.throttler.EndProcessing()

	return nil
}

func (u *userAccountsSyncer) findAllAccountRootHashes(mainTrie data.Trie) ([][]byte, error) {
	leafs, err := mainTrie.GetAllLeaves()
	if err != nil {
		return nil, err
	}

	rootHashes := make([][]byte, 0)
	for _, leaf := range leafs {
		account := state.NewEmptyUserAccount()
		err = u.marshalizer.Unmarshal(account, leaf)
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
			continue
		}

		if len(account.RootHash) > 0 {
			rootHashes = append(rootHashes, account.RootHash)
		}
	}

	return rootHashes, nil
}
