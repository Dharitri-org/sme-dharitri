package coordinator

import (
	"math/big"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data/smartContractResult"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/mock"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
	"github.com/Dharitri-org/sme-vm-common/parsers"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() ArgNewTxTypeHandler {
	return ArgNewTxTypeHandler{
		PubkeyConverter:  createMockPubkeyConverter(),
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(3),
		BuiltInFuncNames: make(map[string]struct{}),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
}

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func TestNewTxTypeHandler_NilAddrConv(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.PubkeyConverter = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewTxTypeHandler_NilShardCoord(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.ShardCoordinator = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewTxTypeHandler_NilArgParser(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.ArgumentParser = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilArgumentParser, err)
}

func TestNewTxTypeHandler_NilBuiltInFuncs(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.BuiltInFuncNames = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilBuiltInFunction, err)
}

func TestNewTxTypeHandler_ValsOk(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)
	assert.False(t, tth.IsInterfaceNil())
}

func TestTxTypeHandler_ComputeTransactionTypeNil(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType := tth.ComputeTransactionType(nil)
	assert.Equal(t, process.InvalidTransaction, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeNilTx(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	tx = nil
	txType := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.InvalidTransaction, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeErrWrongTransaction(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = nil
	tx.Value = big.NewInt(45)

	txType := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.InvalidTransaction, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeScDeployment(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, createMockPubkeyConverter().Len())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	txType := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.SCDeployment, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeRecv0AddressWrongTransaction(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, createMockPubkeyConverter().Len())
	tx.Data = nil
	tx.Value = big.NewInt(45)

	txType := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.InvalidTransaction, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeScInvoking(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255}
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.SCInvoking, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeMoveBalance(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = []byte("001")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.MoveBalance, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeBuiltInFunc(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = []byte("001")
	tx.Data = []byte("builtIn")
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	builtIn := "builtIn"
	arg.BuiltInFuncNames[builtIn] = struct{}{}
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.BuiltInFunctionCall, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeRelayedFunc(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = []byte("001")
	tx.Data = []byte(core.RelayedTransaction)
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.RelayedTx, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeForSCRCallBack(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = []byte("001")
	tx.Data = []byte("00")
	tx.CallType = vmcommon.AsynchronousCallBack
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.SCInvoking, txType)
}
