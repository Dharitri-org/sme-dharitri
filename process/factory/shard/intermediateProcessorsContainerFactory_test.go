package shard_test

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/factory/shard"
	"github.com/Dharitri-org/sme-dharitri/process/mock"
	"github.com/Dharitri-org/sme-dharitri/storage"
	"github.com/Dharitri-org/sme-dharitri/testscommon"
	"github.com/stretchr/testify/assert"
)

func createDataPools() dataRetriever.PoolsHolder {
	pools := testscommon.NewPoolsHolderStub()
	pools.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{}
	}
	pools.MiniBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.PeerChangesBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.MetaBlocksCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.UnsignedTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.RewardTransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
		return testscommon.NewShardedDataStub()
	}
	pools.TrieNodesCalled = func() storage.Cacher {
		return testscommon.NewCacherStub()
	}
	pools.CurrBlockTxsCalled = func() dataRetriever.TransactionCacher {
		return &mock.TxForCurrentBlockStub{}
	}
	return pools
}

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func TestNewIntermediateProcessorsContainerFactory_NilShardCoord(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dPool,
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		nil,
		&mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dPool,
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilHasher(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		nil,
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dPool,
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilAdrConv(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		&mock.ChainStorerMock{},
		dPool,
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewIntermediateProcessorsContainerFactory_NilStorer(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		createMockPubkeyConverter(),
		nil,
		dPool,
	)

	assert.Nil(t, ipcf)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestNewIntermediateProcessorsContainerFactory(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dPool,
	)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)
	assert.False(t, ipcf.IsInterfaceNil())
}

func TestIntermediateProcessorsContainerFactory_Create(t *testing.T) {
	t.Parallel()

	dPool := createDataPools()
	ipcf, err := shard.NewIntermediateProcessorsContainerFactory(
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		createMockPubkeyConverter(),
		&mock.ChainStorerMock{},
		dPool,
	)

	assert.Nil(t, err)
	assert.NotNil(t, ipcf)

	container, err := ipcf.Create()
	assert.Nil(t, err)
	assert.Equal(t, 3, container.Len())
}
