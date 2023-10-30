package factory

import (
	"fmt"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	trieFactory "github.com/Dharitri-org/sme-dharitri/data/trie/factory"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/storage"
)

// TriesComponentsFactoryArgs holds the arguments needed for creating a tries components factory
type TriesComponentsFactoryArgs struct {
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	PathManager      storage.PathManagerHandler
	ShardCoordinator sharding.Coordinator
	Config           config.Config
}

type triesComponentsFactory struct {
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	pathManager      storage.PathManagerHandler
	shardCoordinator sharding.Coordinator
	config           config.Config
}

// NewTriesComponentsFactory return a new instance of tries components factory
func NewTriesComponentsFactory(args TriesComponentsFactoryArgs) (*triesComponentsFactory, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, ErrNilHasher
	}
	if check.IfNil(args.PathManager) {
		return nil, ErrNilPathManager
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}

	return &triesComponentsFactory{
		config:           args.Config,
		marshalizer:      args.Marshalizer,
		hasher:           args.Hasher,
		pathManager:      args.PathManager,
		shardCoordinator: args.ShardCoordinator,
	}, nil
}

// Create creates and returns
func (tcf *triesComponentsFactory) Create() (*TriesComponents, error) {
	trieContainer := state.NewDataTriesHolder()
	trieFactoryArgs := trieFactory.TrieFactoryArgs{
		EvictionWaitingListCfg:   tcf.config.EvictionWaitingList,
		SnapshotDbCfg:            tcf.config.TrieSnapshotDB,
		Marshalizer:              tcf.marshalizer,
		Hasher:                   tcf.hasher,
		PathManager:              tcf.pathManager,
		TrieStorageManagerConfig: tcf.config.TrieStorageManagerConfig,
	}
	shardIDString := convertShardIDToString(tcf.shardCoordinator.SelfId())

	trieFactoryObj, err := trieFactory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, err
	}

	trieStorageManagers := make(map[string]data.StorageManager)
	userStorageManager, userAccountTrie, err := trieFactoryObj.Create(
		tcf.config.AccountsTrieStorage,
		shardIDString,
		tcf.config.StateTriesConfig.AccountsStatePruningEnabled,
		tcf.config.StateTriesConfig.MaxStateTrieLevelInMemory,
	)
	if err != nil {
		return nil, err
	}
	trieContainer.Put([]byte(trieFactory.UserAccountTrie), userAccountTrie)
	trieStorageManagers[trieFactory.UserAccountTrie] = userStorageManager

	peerStorageManager, peerAccountsTrie, err := trieFactoryObj.Create(
		tcf.config.PeerAccountsTrieStorage,
		shardIDString,
		tcf.config.StateTriesConfig.PeerStatePruningEnabled,
		tcf.config.StateTriesConfig.MaxPeerTrieLevelInMemory,
	)
	if err != nil {
		return nil, err
	}
	trieContainer.Put([]byte(trieFactory.PeerAccountTrie), peerAccountsTrie)
	trieStorageManagers[trieFactory.PeerAccountTrie] = peerStorageManager

	return &TriesComponents{
		TriesContainer:      trieContainer,
		TrieStorageManagers: trieStorageManagers,
	}, nil
}

func convertShardIDToString(shardID uint32) string {
	if shardID == core.MetachainShardId {
		return "metachain"
	}

	return fmt.Sprintf("%d", shardID)
}
