package track

import (
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// ArgBaseTracker holds all dependencies required by the process data factory in order to create
// new instances of shard/meta block tracker
type ArgBaseTracker struct {
	Hasher           hashing.Hasher
	HeaderValidator  process.HeaderConstructionValidator
	Marshalizer      marshal.Marshalizer
	RequestHandler   process.RequestHandler
	Rounder          process.Rounder
	ShardCoordinator sharding.Coordinator
	Store            dataRetriever.StorageService
	StartHeaders     map[uint32]data.HeaderHandler
	PoolsHolder      dataRetriever.PoolsHolder
	WhitelistHandler process.WhiteListHandler
}

// ArgShardTracker holds all dependencies required by the process data factory in order to create
// new instances of shard block tracker
type ArgShardTracker struct {
	ArgBaseTracker
}

// ArgMetaTracker holds all dependencies required by the process data factory in order to create
// new instances of meta block tracker
type ArgMetaTracker struct {
	ArgBaseTracker
}
