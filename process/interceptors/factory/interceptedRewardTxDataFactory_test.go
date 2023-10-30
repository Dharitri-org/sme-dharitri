package factory

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/data/rewardTx"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/mock"
	"github.com/Dharitri-org/sme-dharitri/process/rewardTransaction"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedRewardTxDataFactory_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	imh, err := NewInterceptedRewardTxDataFactory(nil)

	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewInterceptedRewardTxDataFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ProtoMarshalizer = nil

	imh, err := NewInterceptedRewardTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedRewardTxDataFactory_NilSignMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.TxSignMarshalizer = nil

	imh, err := NewInterceptedRewardTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedRewardTxDataFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Hasher = nil

	imh, err := NewInterceptedRewardTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedRewardTxDataFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ShardCoordinator = nil

	imh, err := NewInterceptedRewardTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestInterceptedRewardTxDataFactory_NilAdrConverterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.AddressPubkeyConv = nil

	imh, err := NewInterceptedRewardTxDataFactory(arg)
	assert.Nil(t, imh)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestInterceptedRewardTxDataFactory_ShouldWorkAndCreate(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()

	imh, err := NewInterceptedRewardTxDataFactory(arg)
	assert.NotNil(t, imh)
	assert.Nil(t, err)
	assert.False(t, imh.IsInterfaceNil())

	marshalizer := &mock.MarshalizerMock{}
	emptyRewardTx := &rewardTx.RewardTx{}
	emptyRewardTxBuff, _ := marshalizer.Marshal(emptyRewardTx)
	interceptedData, err := imh.Create(emptyRewardTxBuff)
	assert.Nil(t, err)

	_, ok := interceptedData.(*rewardTransaction.InterceptedRewardTransaction)
	assert.True(t, ok)
}
