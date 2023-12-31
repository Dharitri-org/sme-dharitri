package factory

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/core/random"
	"github.com/Dharitri-org/sme-dharitri/core/throttler"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	factoryDataRetriever "github.com/Dharitri-org/sme-dharitri/dataRetriever/factory/resolverscontainer"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/resolvers"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/resolvers/topicResolverSender"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process/factory"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/update"
	"github.com/Dharitri-org/sme-dharitri/update/genesis"
)

const defaultTargetShardID = uint32(0)
const numCrossShardPeers = 2
const numIntraShardPeers = 2

type resolversContainerFactory struct {
	shardCoordinator       sharding.Coordinator
	messenger              dataRetriever.TopicMessageHandler
	marshalizer            marshal.Marshalizer
	intRandomizer          dataRetriever.IntRandomizer
	dataTrieContainer      state.TriesHolder
	container              dataRetriever.ResolversContainer
	intraShardTopic        string
	inputAntifloodHandler  dataRetriever.P2PAntifloodHandler
	outputAntifloodHandler dataRetriever.P2PAntifloodHandler
	throttler              dataRetriever.ResolverThrottler
}

// ArgsNewResolversContainerFactory defines the arguments for the resolversContainerFactory constructor
type ArgsNewResolversContainerFactory struct {
	ShardCoordinator           sharding.Coordinator
	Messenger                  dataRetriever.TopicMessageHandler
	Marshalizer                marshal.Marshalizer
	DataTrieContainer          state.TriesHolder
	ExistingResolvers          dataRetriever.ResolversContainer
	InputAntifloodHandler      dataRetriever.P2PAntifloodHandler
	OutputAntifloodHandler     dataRetriever.P2PAntifloodHandler
	NumConcurrentResolvingJobs int32
}

// NewResolversContainerFactory creates a new container filled with topic resolvers
func NewResolversContainerFactory(args ArgsNewResolversContainerFactory) (*resolversContainerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.Messenger) {
		return nil, update.ErrNilMessenger
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.DataTrieContainer) {
		return nil, update.ErrNilTrieDataGetter
	}
	if check.IfNil(args.ExistingResolvers) {
		return nil, update.ErrNilResolverContainer
	}

	thr, err := throttler.NewNumGoRoutinesThrottler(args.NumConcurrentResolvingJobs)
	if err != nil {
		return nil, err
	}
	intraShardTopic := core.ConsensusTopic +
		args.ShardCoordinator.CommunicationIdentifier(args.ShardCoordinator.SelfId())
	return &resolversContainerFactory{
		shardCoordinator:       args.ShardCoordinator,
		messenger:              args.Messenger,
		marshalizer:            args.Marshalizer,
		intRandomizer:          &random.ConcurrentSafeIntRandomizer{},
		dataTrieContainer:      args.DataTrieContainer,
		container:              args.ExistingResolvers,
		intraShardTopic:        intraShardTopic,
		inputAntifloodHandler:  args.InputAntifloodHandler,
		outputAntifloodHandler: args.OutputAntifloodHandler,
		throttler:              thr,
	}, nil
}

// Create returns a resolver container that will hold all resolvers in the system
func (rcf *resolversContainerFactory) Create() (dataRetriever.ResolversContainer, error) {
	err := rcf.generateTrieNodesResolvers()
	if err != nil {
		return nil, err
	}

	return rcf.container, nil
}

func (rcf *resolversContainerFactory) generateTrieNodesResolvers() error {
	shardC := rcf.shardCoordinator

	keys := make([]string, 0)
	resolversSlice := make([]dataRetriever.Resolver, 0)

	for i := uint32(0); i < shardC.NumberOfShards(); i++ {
		identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(i, core.MetachainShardId)
		if rcf.checkIfResolverExists(identifierTrieNodes) {
			continue
		}

		trieId := genesis.CreateTrieIdentifier(i, genesis.UserAccount)
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, trieId)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	identifierTrieNodes := factory.AccountTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	if !rcf.checkIfResolverExists(identifierTrieNodes) {
		trieId := genesis.CreateTrieIdentifier(core.MetachainShardId, genesis.UserAccount)
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, trieId)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	identifierTrieNodes = factory.ValidatorTrieNodesTopic + core.CommunicationIdentifierBetweenShards(core.MetachainShardId, core.MetachainShardId)
	if !rcf.checkIfResolverExists(identifierTrieNodes) {
		trieID := genesis.CreateTrieIdentifier(core.MetachainShardId, genesis.ValidatorAccount)
		resolver, err := rcf.createTrieNodesResolver(identifierTrieNodes, trieID)
		if err != nil {
			return err
		}

		resolversSlice = append(resolversSlice, resolver)
		keys = append(keys, identifierTrieNodes)
	}

	return rcf.container.AddMultiple(keys, resolversSlice)
}

func (rcf *resolversContainerFactory) checkIfResolverExists(topic string) bool {
	_, err := rcf.container.Get(topic)
	return err == nil
}

func (rcf *resolversContainerFactory) createTrieNodesResolver(baseTopic string, trieId string) (dataRetriever.Resolver, error) {
	peerListCreator, err := topicResolverSender.NewDiffPeerListCreator(
		rcf.messenger,
		baseTopic,
		rcf.intraShardTopic,
		factoryDataRetriever.EmptyExcludePeersOnTopic,
	)
	if err != nil {
		return nil, err
	}

	arg := topicResolverSender.ArgTopicResolverSender{
		Messenger:          rcf.messenger,
		TopicName:          baseTopic,
		PeerListCreator:    peerListCreator,
		Marshalizer:        rcf.marshalizer,
		Randomizer:         rcf.intRandomizer,
		TargetShardId:      defaultTargetShardID,
		OutputAntiflooder:  rcf.outputAntifloodHandler,
		NumCrossShardPeers: numCrossShardPeers,
		NumIntraShardPeers: numIntraShardPeers,
	}
	resolverSender, err := topicResolverSender.NewTopicResolverSender(arg)
	if err != nil {
		return nil, err
	}

	trie := rcf.dataTrieContainer.Get([]byte(trieId))
	argTrieResolver := resolvers.ArgTrieNodeResolver{
		SenderResolver:   resolverSender,
		TrieDataGetter:   trie,
		Marshalizer:      rcf.marshalizer,
		AntifloodHandler: rcf.inputAntifloodHandler,
		Throttler:        rcf.throttler,
	}
	resolver, err := resolvers.NewTrieNodeResolver(argTrieResolver)
	if err != nil {
		return nil, err
	}

	err = rcf.messenger.RegisterMessageProcessor(resolver.RequestTopic(), resolver)
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rcf *resolversContainerFactory) IsInterfaceNil() bool {
	return rcf == nil
}
