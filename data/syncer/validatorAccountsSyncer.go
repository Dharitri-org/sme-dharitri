package syncer

import (
	"context"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/epochStart"
	"github.com/Dharitri-org/sme-dharitri/process/factory"
)

var _ epochStart.AccountsDBSyncer = (*validatorAccountsSyncer)(nil)

type validatorAccountsSyncer struct {
	*baseAccountsSyncer
}

// ArgsNewValidatorAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewValidatorAccountsSyncer struct {
	ArgsNewBaseAccountsSyncer
}

// NewValidatorAccountsSyncer creates a validator account syncer
func NewValidatorAccountsSyncer(args ArgsNewValidatorAccountsSyncer) (*validatorAccountsSyncer, error) {
	err := checkArgs(args.ArgsNewBaseAccountsSyncer)
	if err != nil {
		return nil, err
	}

	b := &baseAccountsSyncer{
		hasher:               args.Hasher,
		marshalizer:          args.Marshalizer,
		trieSyncers:          make(map[string]data.TrieSyncer),
		dataTries:            make(map[string]data.Trie),
		trieStorageManager:   args.TrieStorageManager,
		requestHandler:       args.RequestHandler,
		waitTime:             args.WaitTime,
		shardId:              core.MetachainShardId,
		cacher:               args.Cacher,
		rootHash:             nil,
		maxTrieLevelInMemory: args.MaxTrieLevelInMemory,
	}

	u := &validatorAccountsSyncer{
		baseAccountsSyncer: b,
	}

	return u, nil
}

// SyncAccounts will launch the syncing method to gather all the data needed for validatorAccounts - it is a blocking method
func (v *validatorAccountsSyncer) SyncAccounts(rootHash []byte) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), v.waitTime)
	defer cancel()

	return v.syncMainTrie(rootHash, factory.ValidatorTrieNodesTopic, ctx)
}
