package sposFactory

import (
	"github.com/Dharitri-org/sme-dharitri/consensus"
	"github.com/Dharitri-org/sme-dharitri/consensus/broadcast"
	"github.com/Dharitri-org/sme-dharitri/consensus/spos"
	"github.com/Dharitri-org/sme-dharitri/consensus/spos/bls"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/indexer"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// GetSubroundsFactory returns a subrounds factory depending of the given parameter
func GetSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState *spos.ConsensusState,
	worker spos.WorkerHandler,
	consensusType string,
	appStatusHandler core.AppStatusHandler,
	indexer indexer.Indexer,
	chainID []byte,
	currentPid core.PeerID,
) (spos.SubroundsFactory, error) {
	switch consensusType {
	case blsConsensusType:
		subRoundFactoryBls, err := bls.NewSubroundsFactory(
			consensusDataContainer,
			consensusState,
			worker,
			chainID,
			currentPid,
		)
		if err != nil {
			return nil, err
		}

		err = subRoundFactoryBls.SetAppStatusHandler(appStatusHandler)
		if err != nil {
			return nil, err
		}

		subRoundFactoryBls.SetIndexer(indexer)

		return subRoundFactoryBls, nil
	default:
		return nil, ErrInvalidConsensusType
	}
}

// GetConsensusCoreFactory returns a consensus service depending of the given parameter
func GetConsensusCoreFactory(consensusType string) (spos.ConsensusService, error) {
	switch consensusType {
	case blsConsensusType:
		return bls.NewConsensusService()
	default:
		return nil, ErrInvalidConsensusType
	}
}

// GetBroadcastMessenger returns a consensus service depending of the given parameter
func GetBroadcastMessenger(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	messenger consensus.P2PMessenger,
	shardCoordinator sharding.Coordinator,
	privateKey crypto.PrivateKey,
	peerSignatureHandler crypto.PeerSignatureHandler,
	headersSubscriber consensus.HeadersPoolSubscriber,
	interceptorsContainer process.InterceptorsContainer,
) (consensus.BroadcastMessenger, error) {

	commonMessengerArgs := broadcast.CommonMessengerArgs{
		Marshalizer:                marshalizer,
		Hasher:                     hasher,
		Messenger:                  messenger,
		PrivateKey:                 privateKey,
		ShardCoordinator:           shardCoordinator,
		PeerSignatureHandler:       peerSignatureHandler,
		HeadersSubscriber:          headersSubscriber,
		MaxDelayCacheSize:          maxDelayCacheSize,
		MaxValidatorDelayCacheSize: maxDelayCacheSize,
		InterceptorsContainer:      interceptorsContainer,
	}

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		shardMessengerArgs := broadcast.ShardChainMessengerArgs{
			CommonMessengerArgs: commonMessengerArgs,
		}

		return broadcast.NewShardChainMessenger(shardMessengerArgs)
	}

	if shardCoordinator.SelfId() == core.MetachainShardId {
		metaMessengerArgs := broadcast.MetaChainMessengerArgs{
			CommonMessengerArgs: commonMessengerArgs,
		}

		return broadcast.NewMetaChainMessenger(metaMessengerArgs)
	}

	return nil, ErrInvalidShardId
}
