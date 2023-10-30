package resolverscontainer

import (
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/typeConverters"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// FactoryArgs will hold the arguments for ResolversContainerFactory for both shard and meta
type FactoryArgs struct {
	SizeCheckDelta             uint32
	NumConcurrentResolvingJobs int32
	ShardCoordinator           sharding.Coordinator
	Messenger                  dataRetriever.TopicMessageHandler
	Store                      dataRetriever.StorageService
	Marshalizer                marshal.Marshalizer
	DataPools                  dataRetriever.PoolsHolder
	Uint64ByteSliceConverter   typeConverters.Uint64ByteSliceConverter
	DataPacker                 dataRetriever.DataPacker
	TriesContainer             state.TriesHolder
	InputAntifloodHandler      dataRetriever.P2PAntifloodHandler
	OutputAntifloodHandler     dataRetriever.P2PAntifloodHandler
}
