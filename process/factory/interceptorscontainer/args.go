package interceptorscontainer

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// ShardInterceptorsContainerFactoryArgs holds the arguments needed for ShardInterceptorsContainerFactory
type ShardInterceptorsContainerFactoryArgs struct {
	Accounts                state.AccountsAdapter
	ShardCoordinator        sharding.Coordinator
	NodesCoordinator        sharding.NodesCoordinator
	Messenger               process.TopicHandler
	Store                   dataRetriever.StorageService
	ProtoMarshalizer        marshal.Marshalizer
	TxSignMarshalizer       marshal.Marshalizer
	Hasher                  hashing.Hasher
	KeyGen                  crypto.KeyGenerator
	BlockSignKeyGen         crypto.KeyGenerator
	SingleSigner            crypto.SingleSigner
	BlockSingleSigner       crypto.SingleSigner
	MultiSigner             crypto.MultiSigner
	DataPool                dataRetriever.PoolsHolder
	AddressPubkeyConverter  core.PubkeyConverter
	MaxTxNonceDeltaAllowed  int
	TxFeeHandler            process.FeeHandler
	BlockBlackList          process.TimeCacher
	HeaderSigVerifier       process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier process.InterceptedHeaderIntegrityVerifier
	ValidityAttester        process.ValidityAttester
	EpochStartTrigger       process.EpochStartTriggerHandler
	WhiteListHandler        process.WhiteListHandler
	WhiteListerVerifiedTxs  process.WhiteListHandler
	AntifloodHandler        process.P2PAntifloodHandler
	ArgumentsParser         process.ArgumentsParser
	ChainID                 []byte
	SizeCheckDelta          uint32
	MinTransactionVersion   uint32
}

// MetaInterceptorsContainerFactoryArgs holds the arguments needed for MetaInterceptorsContainerFactory
type MetaInterceptorsContainerFactoryArgs struct {
	ShardCoordinator        sharding.Coordinator
	NodesCoordinator        sharding.NodesCoordinator
	Messenger               process.TopicHandler
	Store                   dataRetriever.StorageService
	ProtoMarshalizer        marshal.Marshalizer
	TxSignMarshalizer       marshal.Marshalizer
	Hasher                  hashing.Hasher
	MultiSigner             crypto.MultiSigner
	DataPool                dataRetriever.PoolsHolder
	Accounts                state.AccountsAdapter
	AddressPubkeyConverter  core.PubkeyConverter
	SingleSigner            crypto.SingleSigner
	BlockSingleSigner       crypto.SingleSigner
	KeyGen                  crypto.KeyGenerator
	BlockKeyGen             crypto.KeyGenerator
	MaxTxNonceDeltaAllowed  int
	TxFeeHandler            process.FeeHandler
	BlackList               process.TimeCacher
	HeaderSigVerifier       process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier process.InterceptedHeaderIntegrityVerifier
	ValidityAttester        process.ValidityAttester
	EpochStartTrigger       process.EpochStartTriggerHandler
	WhiteListHandler        process.WhiteListHandler
	WhiteListerVerifiedTxs  process.WhiteListHandler
	AntifloodHandler        process.P2PAntifloodHandler
	ArgumentsParser         process.ArgumentsParser
	ChainID                 []byte
	MinTransactionVersion   uint32
	SizeCheckDelta          uint32
}
