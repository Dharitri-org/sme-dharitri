package metachain

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/block/postprocess"
	"github.com/Dharitri-org/sme-dharitri/process/factory/containers"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

type intermediateProcessorsContainerFactory struct {
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	pubkeyConverter  core.PubkeyConverter
	store            dataRetriever.StorageService
	poolsHolder      dataRetriever.PoolsHolder
}

// NewIntermediateProcessorsContainerFactory is responsible for creating a new intermediate processors factory object
func NewIntermediateProcessorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	pubkeyConverter core.PubkeyConverter,
	store dataRetriever.StorageService,
	poolsHolder dataRetriever.PoolsHolder,
) (*intermediateProcessorsContainerFactory, error) {

	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(pubkeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(store) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(poolsHolder) {
		return nil, process.ErrNilPoolsHolder
	}

	return &intermediateProcessorsContainerFactory{
		shardCoordinator: shardCoordinator,
		marshalizer:      marshalizer,
		hasher:           hasher,
		pubkeyConverter:  pubkeyConverter,
		poolsHolder:      poolsHolder,
		store:            store,
	}, nil
}

// Create returns a preprocessor container that will hold all preprocessors in the system
func (ppcm *intermediateProcessorsContainerFactory) Create() (process.IntermediateProcessorContainer, error) {
	container := containers.NewIntermediateTransactionHandlersContainer()

	interproc, err := ppcm.createSmartContractResultsIntermediateProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.SmartContractResultBlock, interproc)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ppcm *intermediateProcessorsContainerFactory) createSmartContractResultsIntermediateProcessor() (process.IntermediateTransactionHandler, error) {
	irp, err := postprocess.NewIntermediateResultsProcessor(
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.shardCoordinator,
		ppcm.pubkeyConverter,
		ppcm.store,
		block.SmartContractResultBlock,
		ppcm.poolsHolder.CurrentBlockTxs(),
	)

	return irp, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *intermediateProcessorsContainerFactory) IsInterfaceNil() bool {
	return ppcm == nil
}
