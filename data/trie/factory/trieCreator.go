package factory

import (
	"path"
	"path/filepath"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/trie"
	"github.com/Dharitri-org/sme-dharitri/data/trie/evictionWaitingList"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/storage"
	"github.com/Dharitri-org/sme-dharitri/storage/factory"
	"github.com/Dharitri-org/sme-dharitri/storage/storageUnit"
	logger "github.com/Dharitri-org/sme-logger"
)

type trieCreator struct {
	evictionWaitingListCfg   config.EvictionWaitingListConfig
	snapshotDbCfg            config.DBConfig
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	pathManager              storage.PathManagerHandler
	trieStorageManagerConfig config.TrieStorageManagerConfig
}

var log = logger.GetOrCreate("trie")

// NewTrieFactory creates a new trie factory
func NewTrieFactory(
	args TrieFactoryArgs,
) (*trieCreator, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, trie.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, trie.ErrNilHasher
	}
	if check.IfNil(args.PathManager) {
		return nil, trie.ErrNilPathManager
	}

	return &trieCreator{
		evictionWaitingListCfg:   args.EvictionWaitingListCfg,
		snapshotDbCfg:            args.SnapshotDbCfg,
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		pathManager:              args.PathManager,
		trieStorageManagerConfig: args.TrieStorageManagerConfig,
	}, nil
}

// Create creates a new trie
func (tc *trieCreator) Create(
	trieStorageCfg config.StorageConfig,
	shardID string,
	pruningEnabled bool,
	maxTrieLevelInMem uint,
) (data.StorageManager, data.Trie, error) {
	trieStoragePath, mainDb := path.Split(tc.pathManager.PathForStatic(shardID, trieStorageCfg.DB.FilePath))

	dbConfig := factory.GetDBFromConfig(trieStorageCfg.DB)
	dbConfig.FilePath = path.Join(trieStoragePath, mainDb)
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(trieStorageCfg.Cache),
		dbConfig,
		factory.GetBloomFromConfig(trieStorageCfg.Bloom),
	)
	if err != nil {
		return nil, nil, err
	}

	log.Trace("trie pruning status", "enabled", pruningEnabled)
	if !pruningEnabled {
		trieStorage, errNewTrie := trie.NewTrieStorageManagerWithoutPruning(accountsTrieStorage)
		if errNewTrie != nil {
			return nil, nil, errNewTrie
		}

		newTrie, err := trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher, maxTrieLevelInMem)
		if err != nil {
			return nil, nil, err
		}

		return trieStorage, newTrie, nil
	}

	arg := storageUnit.ArgDB{
		DBType:            storageUnit.DBType(tc.evictionWaitingListCfg.DB.Type),
		Path:              filepath.Join(trieStoragePath, tc.evictionWaitingListCfg.DB.FilePath),
		BatchDelaySeconds: tc.evictionWaitingListCfg.DB.BatchDelaySeconds,
		MaxBatchSize:      tc.evictionWaitingListCfg.DB.MaxBatchSize,
		MaxOpenFiles:      tc.evictionWaitingListCfg.DB.MaxOpenFiles,
	}
	evictionDb, err := storageUnit.NewDB(arg)
	if err != nil {
		return nil, nil, err
	}

	ewl, err := evictionWaitingList.NewEvictionWaitingList(tc.evictionWaitingListCfg.Size, evictionDb, tc.marshalizer)
	if err != nil {
		return nil, nil, err
	}

	snapshotDbCfg := config.DBConfig{
		FilePath:          filepath.Join(trieStoragePath, tc.snapshotDbCfg.FilePath),
		Type:              tc.snapshotDbCfg.Type,
		BatchDelaySeconds: tc.snapshotDbCfg.BatchDelaySeconds,
		MaxBatchSize:      tc.snapshotDbCfg.MaxBatchSize,
		MaxOpenFiles:      tc.snapshotDbCfg.MaxOpenFiles,
	}

	trieStorage, err := trie.NewTrieStorageManager(
		accountsTrieStorage,
		tc.marshalizer,
		tc.hasher,
		snapshotDbCfg,
		ewl,
		tc.trieStorageManagerConfig,
	)
	if err != nil {
		return nil, nil, err
	}

	newTrie, err := trie.NewTrie(trieStorage, tc.marshalizer, tc.hasher, maxTrieLevelInMem)
	if err != nil {
		return nil, nil, err
	}

	return trieStorage, newTrie, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *trieCreator) IsInterfaceNil() bool {
	return tc == nil
}
