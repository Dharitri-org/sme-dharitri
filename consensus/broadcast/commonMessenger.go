package broadcast

import (
	"strings"
	"time"

	"github.com/Dharitri-org/sme-dharitri/consensus"
	"github.com/Dharitri-org/sme-dharitri/consensus/spos"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/core/partitioning"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/factory"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	logger "github.com/Dharitri-org/sme-logger"
)

var log = logger.GetOrCreate("consensus/broadcast")

// delayedBroadcaster exposes functionality for handling the consensus members broadcasting of delay data
type delayedBroadcaster interface {
	SetLeaderData(data *delayedBroadcastData) error
	SetValidatorData(data *delayedBroadcastData) error
	SetHeaderForValidator(vData *validatorHeaderBroadcastData) error
	SetBroadcastHandlers(
		mbBroadcast func(mbData map[uint32][]byte) error,
		txBroadcast func(txData map[string][][]byte) error,
		headerBroadcast func(header data.HeaderHandler) error,
	) error
	Close()
}

type commonMessenger struct {
	marshalizer             marshal.Marshalizer
	hasher                  hashing.Hasher
	messenger               consensus.P2PMessenger
	privateKey              crypto.PrivateKey
	shardCoordinator        sharding.Coordinator
	peerSignatureHandler    crypto.PeerSignatureHandler
	delayedBlockBroadcaster delayedBroadcaster
}

// CommonMessengerArgs holds the arguments for creating commonMessenger instance
type CommonMessengerArgs struct {
	Marshalizer                marshal.Marshalizer
	Hasher                     hashing.Hasher
	Messenger                  consensus.P2PMessenger
	PrivateKey                 crypto.PrivateKey
	ShardCoordinator           sharding.Coordinator
	PeerSignatureHandler       crypto.PeerSignatureHandler
	HeadersSubscriber          consensus.HeadersPoolSubscriber
	InterceptorsContainer      process.InterceptorsContainer
	MaxDelayCacheSize          uint32
	MaxValidatorDelayCacheSize uint32
}

func checkCommonMessengerNilParameters(
	args CommonMessengerArgs,
) error {
	if check.IfNil(args.Marshalizer) {
		return spos.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return spos.ErrNilHasher
	}
	if check.IfNil(args.Messenger) {
		return spos.ErrNilMessenger
	}
	if check.IfNil(args.PrivateKey) {
		return spos.ErrNilPrivateKey
	}
	if check.IfNil(args.ShardCoordinator) {
		return spos.ErrNilShardCoordinator
	}
	if check.IfNil(args.PeerSignatureHandler) {
		return spos.ErrNilPeerSignatureHandler
	}
	if check.IfNil(args.InterceptorsContainer) {
		return spos.ErrNilInterceptorsContainer
	}
	if check.IfNil(args.HeadersSubscriber) {
		return spos.ErrNilHeadersSubscriber
	}
	if args.MaxDelayCacheSize == 0 || args.MaxValidatorDelayCacheSize == 0 {
		return spos.ErrInvalidCacheSize
	}

	return nil
}

// BroadcastConsensusMessage will send on consensus topic the consensus message
func (cm *commonMessenger) BroadcastConsensusMessage(message *consensus.Message) error {
	signature, err := cm.peerSignatureHandler.GetPeerSignature(cm.privateKey, message.OriginatorPid)
	if err != nil {
		return err
	}

	message.Signature = signature

	buff, err := cm.marshalizer.Marshal(message)
	if err != nil {
		return err
	}

	consensusTopic := core.ConsensusTopic +
		cm.shardCoordinator.CommunicationIdentifier(cm.shardCoordinator.SelfId())

	go cm.messenger.Broadcast(consensusTopic, buff)

	return nil
}

// BroadcastMiniBlocks will send on miniblocks topic the cross-shard miniblocks
func (cm *commonMessenger) BroadcastMiniBlocks(miniBlocks map[uint32][]byte) error {
	for k, v := range miniBlocks {
		miniBlocksTopic := factory.MiniBlocksTopic +
			cm.shardCoordinator.CommunicationIdentifier(k)

		go cm.messenger.Broadcast(miniBlocksTopic, v)
	}

	if len(miniBlocks) > 0 {
		log.Debug("sent miniblocks",
			"num minblocks", len(miniBlocks),
		)
	}

	return nil
}

// BroadcastTransactions will send on transaction topic the transactions
func (cm *commonMessenger) BroadcastTransactions(transactions map[string][][]byte) error {
	dataPacker, err := partitioning.NewSimpleDataPacker(cm.marshalizer)
	if err != nil {
		return err
	}

	txs := 0
	var packets [][]byte
	for topic, v := range transactions {
		txs += len(v)
		// forward txs to the destination shards in packets
		packets, err = dataPacker.PackDataInChunks(v, core.MaxBulkTransactionSize)
		if err != nil {
			return err
		}

		for _, buff := range packets {
			go cm.messenger.Broadcast(topic, buff)
		}
	}

	if txs > 0 {
		log.Debug("sent transactions",
			"num txs", txs,
		)
	}

	return nil
}

// BroadcastBlockData broadcasts the miniblocks and transactions
func (cm *commonMessenger) BroadcastBlockData(
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	extraDelayForBroadcast time.Duration,
) {
	time.Sleep(extraDelayForBroadcast)

	if len(miniBlocks) > 0 {
		err := cm.BroadcastMiniBlocks(miniBlocks)
		if err != nil {
			log.Warn("broadcast.BroadcastMiniBlocks", "error", err.Error())
		}
	}

	time.Sleep(core.ExtraDelayBetweenBroadcastMbsAndTxs)

	if len(transactions) > 0 {
		err := cm.BroadcastTransactions(transactions)
		if err != nil {
			log.Warn("broadcast.BroadcastTransactions", "error", err.Error())
		}
	}
}

func (cm *commonMessenger) extractMetaMiniBlocksAndTransactions(
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
) (map[uint32][]byte, map[string][][]byte) {

	metaMiniBlocks := make(map[uint32][]byte, 0)
	metaTransactions := make(map[string][][]byte, 0)

	for shardID, mbsMarshalized := range miniBlocks {
		if shardID != core.MetachainShardId {
			continue
		}

		metaMiniBlocks[shardID] = mbsMarshalized
		delete(miniBlocks, shardID)
	}

	identifier := cm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)

	for broadcastTopic, txsMarshalized := range transactions {
		if !strings.Contains(broadcastTopic, identifier) {
			continue
		}

		metaTransactions[broadcastTopic] = txsMarshalized
		delete(transactions, broadcastTopic)
	}

	return metaMiniBlocks, metaTransactions
}
