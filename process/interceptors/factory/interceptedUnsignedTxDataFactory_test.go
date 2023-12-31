package factory

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/data/smartContractResult"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/mock"
	"github.com/Dharitri-org/sme-dharitri/process/unsigned"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedUnsignedTxDataFactory_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedUnsignedTxDataFactory(nil)

	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ProtoMarshalizer = nil

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilSignMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.TxSignMarshalizer = nil

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Hasher = nil

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ShardCoordinator = nil

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedUnsignedTxDataFactory_NilAddrConverterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.AddressPubkeyConv = nil

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestInterceptedUnsignedTxDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()

	imh, err := NewInterceptedUnsignedTxDataFactory(arg)
	assert.NotNil(t, imh)
	assert.Nil(t, err)
	assert.False(t, imh.IsInterfaceNil())

	marshalizer := &mock.MarshalizerMock{}
	emptyTx := &smartContractResult.SmartContractResult{}
	emptyTxBuff, _ := marshalizer.Marshal(emptyTx)
	interceptedData, err := imh.Create(emptyTxBuff)
	assert.Nil(t, err)

	_, ok := interceptedData.(*unsigned.InterceptedUnsignedTransaction)
	assert.True(t, ok)
}
