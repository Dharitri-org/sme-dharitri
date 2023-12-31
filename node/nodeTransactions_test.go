package node_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/fullHistory"
	"github.com/Dharitri-org/sme-dharitri/data/rewardTx"
	"github.com/Dharitri-org/sme-dharitri/data/smartContractResult"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/node"
	"github.com/Dharitri-org/sme-dharitri/node/mock"
	"github.com/Dharitri-org/sme-dharitri/storage"
	"github.com/Dharitri-org/sme-dharitri/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNode_GetTransaction_InvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()
	_, err := n.GetTransaction("zzz")
	assert.Error(t, err)
}

func TestNode_GetTransaction_ShouldFindInTxCacheAndReturn(t *testing.T) {
	t.Parallel()

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled: getCacherHandler(true, ""),
	}
	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return false
			},
		}),
	)
	expectedTx, _ := getDummyNormalTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx.Nonce, tx.Nonce)
}

func TestNode_GetTransaction_ShouldFindInRwdTxCacheAndReturn(t *testing.T) {
	t.Parallel()

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:       getCacherHandler(false, ""),
		RewardTransactionsCalled: getCacherHandler(true, "reward"),
	}
	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return false
			},
		}),
	)
	expectedTx, _ := getDummyRewardTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(expectedTx.RcvAddr), tx.Receiver)
}

func TestNode_GetTransaction_ShouldFindInUnsignedTxCacheAndReturn(t *testing.T) {
	t.Parallel()

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(true, "unsigned"),
	}
	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return false
			},
		}),
	)
	expectedTx, _ := getUnsignedTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx.Nonce, tx.Nonce)
}

func TestNode_GetTransaction_ShouldFindInTxStorageAndReturn(t *testing.T) {
	t.Parallel()

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return getStorerStub(true)
		},
	}
	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return false
			},
		}),
	)
	expectedTx, _ := getDummyNormalTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx.Nonce, tx.Nonce)
}

func TestNode_GetFullHistoryTransaction(t *testing.T) {
	t.Parallel()

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			// getStorerStub contains a transaction that is created with getDummyNormalTx method
			return getStorerStub(true)
		},
	}

	blockHash := []byte("hash")
	mbHash := []byte("mbHash")
	epoch := uint32(10)
	sndShard := uint32(1)
	rcvShard := uint32(2)
	round := uint64(123)
	blockNonce := uint64(1001)
	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return true
			},
			GetTransactionCalled: func(hash []byte) (*fullHistory.HistoryTransactionWithEpoch, error) {
				return &fullHistory.HistoryTransactionWithEpoch{
					Epoch: epoch,
					TransactionsGroupMetadata: &fullHistory.TransactionsGroupMetadata{
						MbHash:      mbHash,
						HeaderHash:  blockHash,
						HeaderNonce: blockNonce,
						SndShardID:  sndShard,
						RcvShardID:  rcvShard,
						Round:       round,
					},
				}, nil
			},
		}),
	)

	dummyTx, _ := getDummyNormalTx()
	expectedTx := &transaction.ApiTransactionResult{
		Type:       "normal",
		Nonce:      dummyTx.Nonce,
		Round:      round,
		Epoch:      epoch,
		Value:      dummyTx.Value.String(),
		Receiver:   hex.EncodeToString(dummyTx.RcvAddr),
		Sender:     hex.EncodeToString(dummyTx.SndAddr),
		GasPrice:   dummyTx.GasPrice,
		GasLimit:   dummyTx.GasLimit,
		Data:       dummyTx.Data,
		Code:       "",
		Signature:  hex.EncodeToString(dummyTx.Signature),
		SndShard:   sndShard,
		RcvShard:   rcvShard,
		BlockNonce: blockNonce,
		MBHash:     hex.EncodeToString(mbHash),
		BlockHash:  hex.EncodeToString(blockHash),
		Status:     core.TxStatusExecuted,
	}

	// transaction that is returned shoud be the same with expectedTx because
	// expectedTx is formated for a dummyTx( returned by method getDummyNormalTx
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx, tx)
}

func TestNode_GetFullHistoryTransaction_TxInPoolShouldReturnItDirectly(t *testing.T) {
	t.Parallel()

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(true, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return getStorerStub(true)
		},
	}

	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return true
			},
		}),
	)

	dummyTx, _ := getDummyNormalTx()
	expectedTx := &transaction.ApiTransactionResult{
		Type:      "normal",
		Nonce:     dummyTx.Nonce,
		Value:     dummyTx.Value.String(),
		Receiver:  hex.EncodeToString(dummyTx.RcvAddr),
		Sender:    hex.EncodeToString(dummyTx.SndAddr),
		GasPrice:  dummyTx.GasPrice,
		GasLimit:  dummyTx.GasLimit,
		Data:      dummyTx.Data,
		Signature: hex.EncodeToString(dummyTx.Signature),
		Status:    core.TxStatusPartiallyExecuted,
	}

	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx, tx)
}

func TestNode_GetFullHistoryTransaction_TxNotInHistoryStorerShouldErr(t *testing.T) {
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	expectedErr := errors.New("test err")
	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return true
			},
			GetTransactionCalled: func(hash []byte) (*fullHistory.HistoryTransactionWithEpoch, error) {
				return nil, expectedErr
			},
		}),
	)

	tx, err := n.GetTransaction("aaaa")
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, tx)

}

func TestNode_GetTransaction_ShouldFindInRwdTxStorageAndReturn(t *testing.T) {
	t.Parallel()

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			if unitType == dataRetriever.TransactionUnit {
				return getStorerStub(false)
			}

			return getStorerStub(true)
		},
	}
	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return false
			},
		}),
	)
	expectedTx, _ := getDummyNormalTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(expectedTx.RcvAddr), tx.Receiver)
}

func TestNode_GetTransaction_ShouldFindInUnsignedTxStorageAndReturn(t *testing.T) {
	t.Parallel()

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.UnsignedTransactionUnit:
				return getStorerStub(true)
			default:
				return getStorerStub(false)
			}
		},
	}
	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return false
			},
		}),
	)
	expectedTx, _ := getDummyNormalTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx.Nonce, tx.Nonce)
}

func TestNode_GetTransaction_ShouldFindInStorageButErrorUnmarshaling(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error unmarshalling")

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.UnsignedTransactionUnit:
				return getStorerStub(true)
			default:
				return getStorerStub(false)
			}
		},
	}
	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{
			UnmarshalHandler: func(_ interface{}, _ []byte) error {
				return expectedErr
			},
		}, 0),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return false
			},
		}),
	)
	tx, err := n.GetTransaction("aaaa")
	assert.Nil(t, tx)
	assert.Equal(t, expectedErr, err)
}

func TestNode_GetTransaction_ShouldNotFindAndReturnUnknown(t *testing.T) {
	t.Parallel()

	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return getStorerStub(false)
		},
	}
	n, _ := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithHistoryRepository(&testscommon.HistoryProcessorStub{
			IsEnabledCalled: func() bool {
				return false
			},
		}),
	)
	tx, err := n.GetTransaction("aaaa")
	assert.Nil(t, tx)
	assert.Error(t, err)
}

func TestNode_ComputeTransactionStatus(t *testing.T) {
	t.Parallel()

	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return getStorerStub(false)
		},
	}

	shardZeroAddr := []byte("addrShard0")
	shardOneAddr := []byte("addrShard1")
	shardCoordinator := &mock.ShardCoordinatorMock{
		ComputeIdCalled: func(addr []byte) uint32 {
			if bytes.Equal(shardZeroAddr, addr) {
				return 0
			}
			return 1
		},
	}

	n, _ := node.NewNode(
		node.WithDataStore(storer),
		node.WithShardCoordinator(shardCoordinator),
	)

	rwdTxCrossShard := &rewardTx.RewardTx{RcvAddr: shardZeroAddr}
	normalTxIntraShard := &transaction.Transaction{RcvAddr: shardZeroAddr, SndAddr: shardZeroAddr}
	normalTxCrossShard := &transaction.Transaction{RcvAddr: shardOneAddr, SndAddr: shardZeroAddr}
	unsignedTxIntraShard := &smartContractResult.SmartContractResult{RcvAddr: shardZeroAddr, SndAddr: shardZeroAddr}
	unsignedTxCrossShard := &smartContractResult.SmartContractResult{RcvAddr: shardOneAddr, SndAddr: shardZeroAddr}

	// cross shard reward tx in storage source shard
	shardCoordinator.SelfShardId = core.MetachainShardId
	txStatus := n.ComputeTransactionStatus(rwdTxCrossShard, false)
	assert.Equal(t, core.TxStatusPartiallyExecuted, txStatus)

	// cross shard reward tx in pool source shard
	shardCoordinator.SelfShardId = core.MetachainShardId
	txStatus = n.ComputeTransactionStatus(rwdTxCrossShard, true)
	assert.Equal(t, core.TxStatusReceived, txStatus)

	// intra shard transaction in storage
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(normalTxIntraShard, false)
	assert.Equal(t, core.TxStatusExecuted, txStatus)

	// intra shard transaction in pool
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(normalTxIntraShard, true)
	assert.Equal(t, core.TxStatusReceived, txStatus)

	// cross shard transaction in storage source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, false)
	assert.Equal(t, core.TxStatusPartiallyExecuted, txStatus)

	// cross shard transaction in pool source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, true)
	assert.Equal(t, core.TxStatusReceived, txStatus)

	// cross shard transaction in storage destination shard
	shardCoordinator.SelfShardId = 1
	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, false)
	assert.Equal(t, core.TxStatusExecuted, txStatus)

	// cross shard transaction in pool destination shard
	shardCoordinator.SelfShardId = 1
	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, true)
	assert.Equal(t, core.TxStatusPartiallyExecuted, txStatus)

	// intra shard scr in storage source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(unsignedTxIntraShard, false)
	assert.Equal(t, core.TxStatusExecuted, txStatus)

	// intra shard scr in pool source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(unsignedTxIntraShard, true)
	assert.Equal(t, core.TxStatusReceived, txStatus)

	// cross shard scr in storage source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(unsignedTxCrossShard, false)
	assert.Equal(t, core.TxStatusPartiallyExecuted, txStatus)

	// cross shard scr in pool source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(unsignedTxCrossShard, true)
	assert.Equal(t, core.TxStatusReceived, txStatus)
}

func getCacherHandler(find bool, cacherType string) func() dataRetriever.ShardedDataCacherNotifier {
	return func() dataRetriever.ShardedDataCacherNotifier {
		switch cacherType {
		case "reward":
			return &testscommon.ShardedDataStub{
				SearchFirstDataCalled: func(_ []byte) (interface{}, bool) {
					if find {
						tx, _ := getDummyRewardTx()
						return tx, true
					}

					return nil, false
				},
			}
		case "unsigned":
			return &testscommon.ShardedDataStub{
				SearchFirstDataCalled: func(_ []byte) (interface{}, bool) {
					if find {
						tx, _ := getUnsignedTx()
						return tx, true
					}

					return nil, false
				},
			}
		default:
			return &testscommon.ShardedDataStub{
				SearchFirstDataCalled: func(_ []byte) (interface{}, bool) {
					if find {
						tx, _ := getDummyNormalTx()
						return tx, true
					}

					return nil, false
				},
			}
		}
	}
}

func getStorerStub(find bool) storage.Storer {
	return &mock.StorerStub{
		HasCalled: func(_ []byte) error {
			if !find {
				return errors.New("key not found")
			}
			return nil
		},
		SearchFirstCalled: func(_ []byte) ([]byte, error) {
			if !find {
				return nil, errors.New("key not found")
			}
			_, txBytes := getDummyNormalTx()
			return txBytes, nil
		},
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			if !find {
				return nil, errors.New("key not found")
			}
			_, txBytes := getDummyNormalTx()
			return txBytes, nil
		},
	}
}

func getDummyNormalTx() (*transaction.Transaction, []byte) {
	tx := transaction.Transaction{Nonce: 37, RcvAddr: []byte("rcvr")}
	marshalizer := &mock.MarshalizerFake{}
	txBytes, _ := marshalizer.Marshal(&tx)
	return &tx, txBytes
}

func getDummyRewardTx() (*rewardTx.RewardTx, []byte) {
	tx := rewardTx.RewardTx{RcvAddr: []byte("rcvr")}
	marshalizer := &mock.MarshalizerFake{}
	txBytes, _ := marshalizer.Marshal(&tx)
	return &tx, txBytes
}

func getUnsignedTx() (*smartContractResult.SmartContractResult, []byte) {
	tx := smartContractResult.SmartContractResult{RcvAddr: []byte("rcvr")}
	marshalizer := &mock.MarshalizerFake{}
	txBytes, _ := marshalizer.Marshal(&tx)
	return &tx, txBytes
}
