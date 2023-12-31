package process

import (
	"bytes"
	"errors"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/update/mock"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
	"github.com/stretchr/testify/assert"
)

func createMockArgsPendingTransactionProcessor() ArgsPendingTransactionProcessor {
	return ArgsPendingTransactionProcessor{
		Accounts:         &mock.AccountsStub{},
		TxProcessor:      &mock.TxProcessorMock{},
		RwdTxProcessor:   &mock.RewardTxProcessorMock{},
		ScrTxProcessor:   &mock.SCProcessorMock{},
		PubKeyConv:       &mock.PubkeyConverterStub{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
	}
}

func TestNewPendingTransactionProcessor(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingTransactionProcessor()

	pendingTxProcessor, err := NewPendingTransactionProcessor(args)
	assert.NoError(t, err)
	assert.False(t, check.IfNil(pendingTxProcessor))
}

func TestPendingTransactionProcessor_ProcessTransactionsDstMe(t *testing.T) {
	t.Parallel()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")
	addr4 := []byte("addr4")
	addr5 := []byte("addr5")
	args := createMockArgsPendingTransactionProcessor()
	args.PubKeyConv = &mock.PubkeyConverterStub{}

	shardCoordinator := mock.NewOneShardCoordinatorMock()
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if address != nil {
			return 0
		}
		return 1
	}
	_ = shardCoordinator.SetSelfId(1)
	args.ShardCoordinator = shardCoordinator

	args.TxProcessor = &mock.TxProcessorMock{
		ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
			if bytes.Equal(transaction.SndAddr, addr4) {
				return 0, errors.New("localErr")
			}
			return 0, nil
		},
	}

	called := false
	args.Accounts = &mock.AccountsStub{
		CommitCalled: func() ([]byte, error) {
			called = true
			return nil, nil
		},
	}

	pendingTxProcessor, _ := NewPendingTransactionProcessor(args)

	hash1 := "hash1"
	hash2 := "hash2"
	hash3 := "hash3"
	hash4 := "hash4"
	hash5 := "hash5"

	tx1 := &transaction.Transaction{RcvAddr: addr1}
	tx2 := &transaction.Transaction{SndAddr: addr2}
	tx3 := &transaction.Transaction{SndAddr: addr3, RcvAddr: addr3}
	tx4 := &transaction.Transaction{SndAddr: addr4}
	tx5 := &transaction.Transaction{SndAddr: addr5}

	mapTxs := map[string]data.TransactionHandler{
		hash1: tx1,
		hash2: tx2,
		hash3: tx3,
		hash4: tx4,
		hash5: tx5,
	}

	mbSlice, err := pendingTxProcessor.ProcessTransactionsDstMe(mapTxs)
	assert.True(t, called)
	assert.NotNil(t, mbSlice)
	assert.NoError(t, err)
	assert.Equal(t, []byte(hash2), mbSlice[0].TxHashes[0])
}

func TestGetSortedSliceFromMbsMap(t *testing.T) {
	mb1 := &block.MiniBlock{
		SenderShardID: 5,
		Type:          block.TxBlock,
	}
	mb2 := &block.MiniBlock{
		SenderShardID: 5,
		Type:          block.SmartContractResultBlock,
	}
	mb3 := &block.MiniBlock{
		SenderShardID: 2,
	}
	mb4 := &block.MiniBlock{
		SenderShardID: 0,
	}

	mbsMap := map[string]*block.MiniBlock{
		"mb1": mb1, "mb2": mb2, "mb3": mb3, "mb4": mb4,
	}

	expectedSlice := block.MiniBlockSlice{mb4, mb3, mb1, mb2}
	mbsSlice := getSortedSliceFromMbsMap(mbsMap)
	assert.Equal(t, expectedSlice, mbsSlice)
}

func TestRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("rootHash")
	args := createMockArgsPendingTransactionProcessor()
	args.Accounts = &mock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}
	pendingTxProcessor, _ := NewPendingTransactionProcessor(args)

	rHash, _ := pendingTxProcessor.RootHash()
	assert.Equal(t, rootHash, rHash)
}
