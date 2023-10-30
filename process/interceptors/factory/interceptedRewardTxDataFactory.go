package factory

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/rewardTransaction"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

var _ process.InterceptedDataFactory = (*interceptedRewardTxDataFactory)(nil)

type interceptedRewardTxDataFactory struct {
	protoMarshalizer marshal.Marshalizer
	hasher           hashing.Hasher
	pubkeyConverter  core.PubkeyConverter
	shardCoordinator sharding.Coordinator
}

// NewInterceptedRewardTxDataFactory creates an instance of interceptedRewardTxDataFactory
func NewInterceptedRewardTxDataFactory(argument *ArgInterceptedDataFactory) (*interceptedRewardTxDataFactory, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.ProtoMarshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.TxSignMarshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(argument.AddressPubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	return &interceptedRewardTxDataFactory{
		protoMarshalizer: argument.ProtoMarshalizer,
		hasher:           argument.Hasher,
		pubkeyConverter:  argument.AddressPubkeyConv,
		shardCoordinator: argument.ShardCoordinator,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (irtdf *interceptedRewardTxDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	return rewardTransaction.NewInterceptedRewardTransaction(
		buff,
		irtdf.protoMarshalizer,
		irtdf.hasher,
		irtdf.pubkeyConverter,
		irtdf.shardCoordinator,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (irtdf *interceptedRewardTxDataFactory) IsInterfaceNil() bool {
	return irtdf == nil
}
