package factory

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/unsigned"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

var _ process.InterceptedDataFactory = (*interceptedUnsignedTxDataFactory)(nil)

type interceptedUnsignedTxDataFactory struct {
	protoMarshalizer marshal.Marshalizer
	hasher           hashing.Hasher
	pubkeyConverter  core.PubkeyConverter
	shardCoordinator sharding.Coordinator
}

// NewInterceptedUnsignedTxDataFactory creates an instance of interceptedUnsignedTxDataFactory
func NewInterceptedUnsignedTxDataFactory(argument *ArgInterceptedDataFactory) (*interceptedUnsignedTxDataFactory, error) {
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

	return &interceptedUnsignedTxDataFactory{
		protoMarshalizer: argument.ProtoMarshalizer,
		hasher:           argument.Hasher,
		pubkeyConverter:  argument.AddressPubkeyConv,
		shardCoordinator: argument.ShardCoordinator,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (iutdf *interceptedUnsignedTxDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	return unsigned.NewInterceptedUnsignedTransaction(
		buff,
		iutdf.protoMarshalizer,
		iutdf.hasher,
		iutdf.pubkeyConverter,
		iutdf.shardCoordinator,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (iutdf *interceptedUnsignedTxDataFactory) IsInterfaceNil() bool {
	return iutdf == nil
}
