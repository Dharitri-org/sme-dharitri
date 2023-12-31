package process

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/update"
)

// TODO: save blocks and transactions to storage
// TODO: marshalize if there are cross shard results (smart contract results in special)
// TODO: use new genesis block process, integrate with it

// ArgsAfterHardFork defines the arguments for the new after hard fork process handler
type ArgsAfterHardFork struct {
	MapBlockProcessors map[uint32]update.HardForkBlockProcessor
	ImportHandler      update.ImportHandler
	ShardCoordinator   sharding.Coordinator
	Hasher             hashing.Hasher
	Marshalizer        marshal.Marshalizer
}

type afterHardFork struct {
	mapBlockProcessors map[uint32]update.HardForkBlockProcessor
	importHandler      update.ImportHandler
	shardCoordinator   sharding.Coordinator
	hasher             hashing.Hasher
	marshalizer        marshal.Marshalizer
}

// NewAfterHardForkBlockCreation creates the after hard fork block creator process handler
func NewAfterHardForkBlockCreation(args ArgsAfterHardFork) (*afterHardFork, error) {
	if args.MapBlockProcessors == nil {
		return nil, update.ErrNilHardForkBlockProcessor
	}
	if check.IfNil(args.ImportHandler) {
		return nil, update.ErrNilImportHandler
	}
	if check.IfNil(args.Hasher) {
		return nil, update.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}

	return &afterHardFork{
		mapBlockProcessors: args.MapBlockProcessors,
		importHandler:      args.ImportHandler,
		shardCoordinator:   args.ShardCoordinator,
		hasher:             args.Hasher,
		marshalizer:        args.Marshalizer,
	}, nil
}

// CreateAllBlocksAfterHardfork creates all the blocks after hardfork
func (a *afterHardFork) CreateAllBlocksAfterHardfork(
	chainID string,
	round uint64,
	nonce uint64,
	epoch uint32,
) (map[uint32]data.HeaderHandler, map[uint32]data.BodyHandler, error) {
	mapHeaders := make(map[uint32]data.HeaderHandler)
	mapBodies := make(map[uint32]data.BodyHandler)

	shardIDs := make([]uint32, a.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < a.shardCoordinator.NumberOfShards(); i++ {
		shardIDs[i] = i
	}
	shardIDs[a.shardCoordinator.NumberOfShards()] = core.MetachainShardId

	for _, shardID := range shardIDs {
		blockProcessor, ok := a.mapBlockProcessors[shardID]
		if !ok {
			return nil, nil, update.ErrNilHardForkBlockProcessor
		}

		hdr, body, err := blockProcessor.CreateNewBlock(chainID, round, nonce, epoch)
		if err != nil {
			return nil, nil, err
		}

		mapHeaders[shardID] = hdr
		mapBodies[shardID] = body
	}

	return mapHeaders, mapBodies, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (a *afterHardFork) IsInterfaceNil() bool {
	return a == nil
}
