package factory

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/trie"
	factoryTrie "github.com/Dharitri-org/sme-dharitri/process/factory"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/storage"
	"github.com/Dharitri-org/sme-dharitri/update"
	containers "github.com/Dharitri-org/sme-dharitri/update/container"
	"github.com/Dharitri-org/sme-dharitri/update/genesis"
)

// ArgsNewTrieSyncersContainerFactory defines the arguments needed to create trie syncers container
type ArgsNewTrieSyncersContainerFactory struct {
	TrieCacher        storage.Cacher
	SyncFolder        string
	RequestHandler    update.RequestHandler
	DataTrieContainer state.TriesHolder
	ShardCoordinator  sharding.Coordinator
}

type trieSyncersContainerFactory struct {
	shardCoordinator sharding.Coordinator
	trieCacher       storage.Cacher
	trieContainer    state.TriesHolder
	requestHandler   update.RequestHandler
}

// NewTrieSyncersContainerFactory creates a factory for trie syncers container
func NewTrieSyncersContainerFactory(args ArgsNewTrieSyncersContainerFactory) (*trieSyncersContainerFactory, error) {
	if len(args.SyncFolder) < 2 {
		return nil, update.ErrInvalidFolderName
	}
	if check.IfNil(args.RequestHandler) {
		return nil, update.ErrNilRequestHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.DataTrieContainer) {
		return nil, update.ErrNilDataTrieContainer
	}
	if check.IfNil(args.TrieCacher) {
		return nil, update.ErrNilCacher
	}

	t := &trieSyncersContainerFactory{
		shardCoordinator: args.ShardCoordinator,
		trieCacher:       args.TrieCacher,
		requestHandler:   args.RequestHandler,
		trieContainer:    args.DataTrieContainer,
	}

	return t, nil
}

// Create creates all the needed syncers and returns the container
func (t *trieSyncersContainerFactory) Create() (update.TrieSyncContainer, error) {
	container := containers.NewTrieSyncersContainer()

	for i := uint32(0); i < t.shardCoordinator.NumberOfShards(); i++ {
		err := t.createOneTrieSyncer(i, genesis.UserAccount, container)
		if err != nil {
			return nil, err
		}
	}

	err := t.createOneTrieSyncer(core.MetachainShardId, genesis.UserAccount, container)
	if err != nil {
		return nil, err
	}

	err = t.createOneTrieSyncer(core.MetachainShardId, genesis.ValidatorAccount, container)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (t *trieSyncersContainerFactory) createOneTrieSyncer(
	shId uint32,
	accType genesis.Type,
	container update.TrieSyncContainer,
) error {
	trieId := genesis.CreateTrieIdentifier(shId, accType)

	dataTrie := t.trieContainer.Get([]byte(trieId))
	if check.IfNil(dataTrie) {
		return update.ErrNilDataTrieContainer
	}

	trieSyncer, err := trie.NewTrieSyncer(t.requestHandler, t.trieCacher, dataTrie, shId, trieTopicFromAccountType(accType))
	if err != nil {
		return err
	}

	err = container.Add(trieId, trieSyncer)
	if err != nil {
		return err
	}

	return nil
}

func trieTopicFromAccountType(accType genesis.Type) string {
	switch accType {
	case genesis.UserAccount:
		return factoryTrie.AccountTrieNodesTopic
	case genesis.ValidatorAccount:
		return factoryTrie.ValidatorTrieNodesTopic
	}
	return ""
}

// IsInterfaceNil returns true if the underlying object is nil
func (t *trieSyncersContainerFactory) IsInterfaceNil() bool {
	return t == nil
}
