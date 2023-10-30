package bootstrap

import (
	"context"
	"time"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/typeConverters"
	"github.com/Dharitri-org/sme-dharitri/epochStart"
	"github.com/Dharitri-org/sme-dharitri/epochStart/bootstrap/disabled"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/economics"
	"github.com/Dharitri-org/sme-dharitri/process/factory"
	"github.com/Dharitri-org/sme-dharitri/process/headerCheck"
	"github.com/Dharitri-org/sme-dharitri/process/interceptors"
	interceptorsFactory "github.com/Dharitri-org/sme-dharitri/process/interceptors/factory"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

var _ epochStart.StartOfEpochMetaSyncer = (*epochStartMetaSyncer)(nil)

type epochStartMetaSyncer struct {
	requestHandler        RequestHandler
	messenger             Messenger
	marshalizer           marshal.Marshalizer
	hasher                hashing.Hasher
	singleDataInterceptor process.Interceptor
	metaBlockProcessor    EpochStartMetaBlockInterceptorProcessor
}

// ArgsNewEpochStartMetaSyncer -
type ArgsNewEpochStartMetaSyncer struct {
	RequestHandler     RequestHandler
	Messenger          Messenger
	Marshalizer        marshal.Marshalizer
	TxSignMarshalizer  marshal.Marshalizer
	ShardCoordinator   sharding.Coordinator
	KeyGen             crypto.KeyGenerator
	BlockKeyGen        crypto.KeyGenerator
	Hasher             hashing.Hasher
	Signer             crypto.SingleSigner
	BlockSigner        crypto.SingleSigner
	ChainID            []byte
	EconomicsData      *economics.EconomicsData
	WhitelistHandler   process.WhiteListHandler
	AddressPubkeyConv  core.PubkeyConverter
	NonceConverter     typeConverters.Uint64ByteSliceConverter
	StartInEpochConfig config.EpochStartConfig
	ArgsParser         process.ArgumentsParser
}

// thresholdForConsideringMetaBlockCorrect represents the percentage (between 0 and 100) of connected peers to send
// the same meta block in order to consider it correct
const thresholdForConsideringMetaBlockCorrect = 67

// NewEpochStartMetaSyncer will return a new instance of epochStartMetaSyncer
func NewEpochStartMetaSyncer(args ArgsNewEpochStartMetaSyncer) (*epochStartMetaSyncer, error) {
	if check.IfNil(args.AddressPubkeyConv) {
		return nil, epochStart.ErrNilPubkeyConverter
	}

	e := &epochStartMetaSyncer{
		requestHandler: args.RequestHandler,
		messenger:      args.Messenger,
		marshalizer:    args.Marshalizer,
		hasher:         args.Hasher,
	}

	processor, err := NewEpochStartMetaBlockProcessor(
		args.Messenger,
		args.RequestHandler,
		args.Marshalizer,
		args.Hasher,
		thresholdForConsideringMetaBlockCorrect,
		args.StartInEpochConfig.MinNumConnectedPeersToStart,
		args.StartInEpochConfig.MinNumOfPeersToConsiderBlockValid,
	)
	if err != nil {
		return nil, err
	}
	e.metaBlockProcessor = processor
	headerIntegrityVerifier, err := headerCheck.NewHeaderIntegrityVerifier(args.ChainID)
	if err != nil {
		return nil, err
	}

	argsInterceptedDataFactory := interceptorsFactory.ArgInterceptedDataFactory{
		ProtoMarshalizer:        args.Marshalizer,
		TxSignMarshalizer:       args.TxSignMarshalizer,
		Hasher:                  args.Hasher,
		ShardCoordinator:        args.ShardCoordinator,
		MultiSigVerifier:        disabled.NewMultiSigVerifier(),
		NodesCoordinator:        disabled.NewNodesCoordinator(),
		KeyGen:                  args.KeyGen,
		BlockKeyGen:             args.BlockKeyGen,
		Signer:                  args.Signer,
		BlockSigner:             args.BlockSigner,
		AddressPubkeyConv:       args.AddressPubkeyConv,
		FeeHandler:              args.EconomicsData,
		HeaderSigVerifier:       disabled.NewHeaderSigVerifier(),
		HeaderIntegrityVerifier: headerIntegrityVerifier,
		ValidityAttester:        disabled.NewValidityAttester(),
		EpochStartTrigger:       disabled.NewEpochStartTrigger(),
		ArgsParser:              args.ArgsParser,
	}

	interceptedMetaHdrDataFactory, err := interceptorsFactory.NewInterceptedMetaHeaderDataFactory(&argsInterceptedDataFactory)
	if err != nil {
		return nil, err
	}

	e.singleDataInterceptor, err = interceptors.NewSingleDataInterceptor(
		factory.MetachainBlocksTopic,
		interceptedMetaHdrDataFactory,
		processor,
		disabled.NewThrottler(),
		disabled.NewAntiFloodHandler(),
		args.WhitelistHandler,
	)
	if err != nil {
		return nil, err
	}

	return e, nil
}

// SyncEpochStartMeta syncs the latest epoch start metablock
func (e *epochStartMetaSyncer) SyncEpochStartMeta(timeToWait time.Duration) (*block.MetaBlock, error) {
	err := e.initTopicForEpochStartMetaBlockInterceptor()
	if err != nil {
		return nil, err
	}
	defer func() {
		e.resetTopicsAndInterceptors()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeToWait)
	mb, errConsensusNotReached := e.metaBlockProcessor.GetEpochStartMetaBlock(ctx)
	cancel()

	if errConsensusNotReached != nil {
		return nil, errConsensusNotReached
	}

	return mb, nil
}

func (e *epochStartMetaSyncer) resetTopicsAndInterceptors() {
	err := e.messenger.UnregisterMessageProcessor(factory.MetachainBlocksTopic)
	if err != nil {
		log.Trace("error unregistering message processors", "error", err)
	}
}

func (e *epochStartMetaSyncer) initTopicForEpochStartMetaBlockInterceptor() error {
	err := e.messenger.CreateTopic(factory.MetachainBlocksTopic, true)
	if err != nil {
		log.Warn("error messenger create topic", "error", err)
		return err
	}

	e.resetTopicsAndInterceptors()
	err = e.messenger.RegisterMessageProcessor(factory.MetachainBlocksTopic, e.singleDataInterceptor)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (e *epochStartMetaSyncer) IsInterfaceNil() bool {
	return e == nil
}
