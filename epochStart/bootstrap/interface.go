package bootstrap

import (
	"context"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// StartOfEpochNodesConfigHandler defines the methods to process nodesConfig from epoch start metablocks
type StartOfEpochNodesConfigHandler interface {
	NodesConfigFromMetaBlock(currMetaBlock *block.MetaBlock, prevMetaBlock *block.MetaBlock) (*sharding.NodesCoordinatorRegistry, uint32, error)
	IsInterfaceNil() bool
}

// EpochStartMetaBlockInterceptorProcessor defines the methods to sync an epoch start metablock
type EpochStartMetaBlockInterceptorProcessor interface {
	process.InterceptorProcessor
	GetEpochStartMetaBlock(ctx context.Context) (*block.MetaBlock, error)
}

// StartInEpochNodesCoordinator defines the methods to process and save nodesCoordinator information to storage
type StartInEpochNodesCoordinator interface {
	EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	NodesCoordinatorToRegistry() *sharding.NodesCoordinatorRegistry
	ShardIdForEpoch(epoch uint32) (uint32, error)
	IsInterfaceNil() bool
}

// Messenger defines which methods a p2p messenger should implement
type Messenger interface {
	dataRetriever.MessageHandler
	dataRetriever.TopicHandler
	UnregisterMessageProcessor(topic string) error
	UnregisterAllMessageProcessors() error
	UnjoinAllTopics() error
	ConnectedPeers() []core.PeerID
}

// RequestHandler defines which methods a request handler should implement
type RequestHandler interface {
	RequestStartOfEpochMetaBlock(epoch uint32)
	SetNumPeersToQuery(topic string, intra int, cross int) error
	GetNumPeersToQuery(topic string) (int, int, error)
	IsInterfaceNil() bool
}
