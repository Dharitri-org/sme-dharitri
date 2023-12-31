package factory

import (
	"fmt"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/data/typeConverters/uint64ByteSlice"
	factoryHasher "github.com/Dharitri-org/sme-dharitri/hashing/factory"
	factoryMarshalizer "github.com/Dharitri-org/sme-dharitri/marshal/factory"
	"github.com/Dharitri-org/sme-dharitri/statusHandler"
)

// CoreComponentsFactoryArgs holds the arguments needed for creating a core components factory
type CoreComponentsFactoryArgs struct {
	Config                config.Config
	ShardId               string
	ChainID               []byte
	MinTransactionVersion uint32
}

// CoreComponentsFactory is responsible for creating the core components
type CoreComponentsFactory struct {
	config                config.Config
	shardId               string
	chainID               []byte
	MinTransactionVersion uint32
}

// NewCoreComponentsFactory initializes the factory which is responsible to creating core components
func NewCoreComponentsFactory(args CoreComponentsFactoryArgs) *CoreComponentsFactory {
	return &CoreComponentsFactory{
		config:                args.Config,
		shardId:               args.ShardId,
		chainID:               args.ChainID,
		MinTransactionVersion: args.MinTransactionVersion,
	}
}

// Create creates the core components
func (ccf *CoreComponentsFactory) Create() (*CoreComponents, error) {
	hasher, err := factoryHasher.NewHasher(ccf.config.Hasher.Type)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrHasherCreation, err.Error())
	}

	internalMarshalizer, err := factoryMarshalizer.NewMarshalizer(ccf.config.Marshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (internal): %s", ErrMarshalizerCreation, err.Error())
	}

	vmMarshalizer, err := factoryMarshalizer.NewMarshalizer(ccf.config.VmMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (vm): %s", ErrMarshalizerCreation, err.Error())
	}

	txSignMarshalizer, err := factoryMarshalizer.NewMarshalizer(ccf.config.TxSignMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (tx sign): %s", ErrMarshalizerCreation, err.Error())
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	return &CoreComponents{
		Hasher:                   hasher,
		InternalMarshalizer:      internalMarshalizer,
		VmMarshalizer:            vmMarshalizer,
		TxSignMarshalizer:        txSignMarshalizer,
		Uint64ByteSliceConverter: uint64ByteSliceConverter,
		StatusHandler:            statusHandler.NewNilStatusHandler(),
		ChainID:                  ccf.chainID,
		MinTransactionVersion:    ccf.MinTransactionVersion,
	}, nil
}
