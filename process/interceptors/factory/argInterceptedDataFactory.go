package factory

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// ArgInterceptedDataFactory holds all dependencies required by the shard and meta intercepted data factory in order to create
// new instances
type ArgInterceptedDataFactory struct {
	ProtoMarshalizer        marshal.Marshalizer
	TxSignMarshalizer       marshal.Marshalizer
	Hasher                  hashing.Hasher
	ShardCoordinator        sharding.Coordinator
	MultiSigVerifier        crypto.MultiSigVerifier
	NodesCoordinator        sharding.NodesCoordinator
	KeyGen                  crypto.KeyGenerator
	BlockKeyGen             crypto.KeyGenerator
	Signer                  crypto.SingleSigner
	BlockSigner             crypto.SingleSigner
	AddressPubkeyConv       core.PubkeyConverter
	FeeHandler              process.FeeHandler
	WhiteListerVerifiedTxs  process.WhiteListHandler
	HeaderSigVerifier       process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier process.InterceptedHeaderIntegrityVerifier
	ValidityAttester        process.ValidityAttester
	EpochStartTrigger       process.EpochStartTriggerHandler
	ArgsParser              process.ArgumentsParser
	ChainID                 []byte
	MinTransactionVersion   uint32
}
