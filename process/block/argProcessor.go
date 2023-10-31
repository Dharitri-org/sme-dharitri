package block

import (
	"github.com/Dharitri-org/sme-dharitri/consensus"
	"github.com/Dharitri-org/sme-dharitri/core/fullHistory"
	"github.com/Dharitri-org/sme-dharitri/core/indexer"
	"github.com/Dharitri-org/sme-dharitri/core/statistics"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/typeConverters"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/node/external"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// ArgBaseProcessor holds all dependencies required by the process data factory in order to create
// new instances
type ArgBaseProcessor struct {
	AccountsDB             map[state.AccountsDbIdentifier]state.AccountsAdapter
	ForkDetector           process.ForkDetector
	Hasher                 hashing.Hasher
	Marshalizer            marshal.Marshalizer
	Store                  dataRetriever.StorageService
	ShardCoordinator       sharding.Coordinator
	NodesCoordinator       sharding.NodesCoordinator
	FeeHandler             process.TransactionFeeHandler
	Uint64Converter        typeConverters.Uint64ByteSliceConverter
	RequestHandler         process.RequestHandler
	BlockChainHook         process.BlockChainHookHandler
	TxCoordinator          process.TransactionCoordinator
	EpochStartTrigger      process.EpochStartTriggerHandler
	HeaderValidator        process.HeaderConstructionValidator
	Rounder                consensus.Rounder
	BootStorer             process.BootStorer
	BlockTracker           process.BlockTracker
	DataPool               dataRetriever.PoolsHolder
	BlockChain             data.ChainHandler
	StateCheckpointModulus uint
	BlockSizeThrottler     process.BlockSizeThrottler
	Indexer                indexer.Indexer
	TpsBenchmark           statistics.TPSBenchmark
	Version                string
	HistoryRepository      fullHistory.HistoryRepository
}

// ArgShardProcessor holds all dependencies required by the process data factory in order to create
// new instances of shard processor
type ArgShardProcessor struct {
	ArgBaseProcessor
}

// ArgMetaProcessor holds all dependencies required by the process data factory in order to create
// new instances of meta processor
type ArgMetaProcessor struct {
	ArgBaseProcessor
	PendingMiniBlocksHandler     process.PendingMiniBlocksHandler
	SCDataGetter                 external.SCQueryService
	SCToProtocol                 process.SmartContractToProtocolHandler
	EpochStartDataCreator        process.EpochStartDataCreator
	EpochEconomics               process.EndOfEpochEconomics
	EpochRewardsCreator          process.EpochStartRewardsCreator
	EpochValidatorInfoCreator    process.EpochStartValidatorInfoCreator
	ValidatorStatisticsProcessor process.ValidatorStatisticsProcessor
}
