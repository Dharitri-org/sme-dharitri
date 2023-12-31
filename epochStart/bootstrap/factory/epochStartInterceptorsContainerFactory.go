package factory

import (
	"time"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/data/typeConverters"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/epochStart"
	"github.com/Dharitri-org/sme-dharitri/epochStart/bootstrap/disabled"
	disabledGenesis "github.com/Dharitri-org/sme-dharitri/genesis/process/disabled"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/factory/interceptorscontainer"
	"github.com/Dharitri-org/sme-dharitri/process/headerCheck"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/storage/timecache"
	"github.com/Dharitri-org/sme-dharitri/update"
)

const timeSpanForBadHeaders = time.Minute

// ArgsEpochStartInterceptorContainer holds the arguments needed for creating a new epoch start interceptors
// container factory
type ArgsEpochStartInterceptorContainer struct {
	Config                 config.Config
	ShardCoordinator       sharding.Coordinator
	TxSignMarshalizer      marshal.Marshalizer
	ProtoMarshalizer       marshal.Marshalizer
	Hasher                 hashing.Hasher
	Messenger              process.TopicHandler
	DataPool               dataRetriever.PoolsHolder
	SingleSigner           crypto.SingleSigner
	BlockSingleSigner      crypto.SingleSigner
	KeyGen                 crypto.KeyGenerator
	BlockKeyGen            crypto.KeyGenerator
	WhiteListHandler       update.WhiteListHandler
	WhiteListerVerifiedTxs update.WhiteListHandler
	AddressPubkeyConv      core.PubkeyConverter
	NonceConverter         typeConverters.Uint64ByteSliceConverter
	ChainID                []byte
	ArgumentsParser        process.ArgumentsParser
	MinTransactionVersion  uint32
}

// NewEpochStartInterceptorsContainer will return a real interceptors container factory, but with many disabled components
func NewEpochStartInterceptorsContainer(args ArgsEpochStartInterceptorContainer) (process.InterceptorsContainer, error) {
	nodesCoordinator := disabled.NewNodesCoordinator()
	storer := disabled.NewChainStorer()
	antiFloodHandler := disabled.NewAntiFloodHandler()
	multiSigner := disabled.NewMultiSigner()
	accountsAdapter := disabled.NewAccountsAdapter()
	if check.IfNil(args.AddressPubkeyConv) {
		return nil, epochStart.ErrNilPubkeyConverter
	}
	blackListHandler := timecache.NewTimeCache(timeSpanForBadHeaders)
	feeHandler := &disabledGenesis.FeeHandler{}
	headerSigVerifier := disabled.NewHeaderSigVerifier()
	headerIntegrityVerifier, err := headerCheck.NewHeaderIntegrityVerifier(args.ChainID)
	if err != nil {
		return nil, err
	}
	sizeCheckDelta := 0
	validityAttester := disabled.NewValidityAttester()
	epochStartTrigger := disabled.NewEpochStartTrigger()

	containerFactoryArgs := interceptorscontainer.MetaInterceptorsContainerFactoryArgs{
		ShardCoordinator:        args.ShardCoordinator,
		NodesCoordinator:        nodesCoordinator,
		Messenger:               args.Messenger,
		Store:                   storer,
		ProtoMarshalizer:        args.ProtoMarshalizer,
		TxSignMarshalizer:       args.TxSignMarshalizer,
		Hasher:                  args.Hasher,
		MultiSigner:             multiSigner,
		DataPool:                args.DataPool,
		Accounts:                accountsAdapter,
		AddressPubkeyConverter:  args.AddressPubkeyConv,
		SingleSigner:            args.SingleSigner,
		BlockSingleSigner:       args.BlockSingleSigner,
		KeyGen:                  args.KeyGen,
		BlockKeyGen:             args.BlockKeyGen,
		MaxTxNonceDeltaAllowed:  core.MaxTxNonceDeltaAllowed,
		TxFeeHandler:            feeHandler,
		BlackList:               blackListHandler,
		HeaderSigVerifier:       headerSigVerifier,
		HeaderIntegrityVerifier: headerIntegrityVerifier,
		SizeCheckDelta:          uint32(sizeCheckDelta),
		ValidityAttester:        validityAttester,
		EpochStartTrigger:       epochStartTrigger,
		WhiteListHandler:        args.WhiteListHandler,
		WhiteListerVerifiedTxs:  args.WhiteListerVerifiedTxs,
		AntifloodHandler:        antiFloodHandler,
		ArgumentsParser:         args.ArgumentsParser,
		ChainID:                 args.ChainID,
		MinTransactionVersion:   args.MinTransactionVersion,
	}

	interceptorsContainerFactory, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(containerFactoryArgs)
	if err != nil {
		return nil, err
	}

	container, err := interceptorsContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	err = interceptorsContainerFactory.AddShardTrieNodeInterceptors(container)
	if err != nil {
		return nil, err
	}

	return container, nil
}
