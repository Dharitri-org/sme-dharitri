package sync

import (
	"time"

	"github.com/Dharitri-org/sme-dharitri/consensus"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/typeConverters"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// ArgBaseBootstrapper holds all dependencies required by the bootstrap data factory in order to create
// new instances
type ArgBaseBootstrapper struct {
	PoolsHolder         dataRetriever.PoolsHolder
	Store               dataRetriever.StorageService
	ChainHandler        data.ChainHandler
	Rounder             consensus.Rounder
	BlockProcessor      process.BlockProcessor
	WaitTime            time.Duration
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	ForkDetector        process.ForkDetector
	RequestHandler      process.RequestHandler
	ShardCoordinator    sharding.Coordinator
	Accounts            state.AccountsAdapter
	BlackListHandler    process.TimeCacher
	NetworkWatcher      process.NetworkConnectionWatcher
	BootStorer          process.BootStorer
	StorageBootstrapper process.BootstrapperFromStorage
	EpochHandler        dataRetriever.EpochHandler
	MiniblocksProvider  process.MiniBlockProvider
	Uint64Converter     typeConverters.Uint64ByteSliceConverter
}

// ArgShardBootstrapper holds all dependencies required by the bootstrap data factory in order to create
// new instances of shard bootstrapper
type ArgShardBootstrapper struct {
	ArgBaseBootstrapper
}

// ArgMetaBootstrapper holds all dependencies required by the bootstrap data factory in order to create
// new instances of meta bootstrapper
type ArgMetaBootstrapper struct {
	ArgBaseBootstrapper
	EpochBootstrapper process.EpochBootstrapper
}
