package smartContract

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/smartContractResult"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/mock"
	"github.com/Dharitri-org/sme-dharitri/process/smartContract/builtInFunctions"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
	"github.com/Dharitri-org/sme-vm-common/parsers"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func generateEmptyByteSlice(size int) []byte {
	buff := make([]byte, size)

	return buff
}

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func createAccounts(tx *transaction.Transaction) (state.UserAccountHandler, state.UserAccountHandler) {
	acntSrc, _ := state.NewUserAccount(tx.SndAddr)
	acntSrc.Balance = acntSrc.Balance.Add(acntSrc.Balance, tx.Value)
	totalFee := big.NewInt(0)
	totalFee = totalFee.Mul(big.NewInt(int64(tx.GasLimit)), big.NewInt(int64(tx.GasPrice)))
	acntSrc.Balance.Set(acntSrc.Balance.Add(acntSrc.Balance, totalFee))

	acntDst, _ := state.NewUserAccount(tx.RcvAddr)

	return acntSrc, acntDst
}

func createMockSmartContractProcessorArguments() ArgsNewSmartContractProcessor {
	return ArgsNewSmartContractProcessor{
		VmContainer: &mock.VMContainerMock{},
		ArgsParser:  &mock.ArgumentParserMock{},
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
		AccountsDB: &mock.AccountsStub{
			RevertToSnapshotCalled: func(snapshot int) error {
				return nil
			},
		},
		BlockChainHook:  &mock.BlockChainHookHandlerMock{},
		PubkeyConv:      createMockPubkeyConverter(),
		Coordinator:     mock.NewMultiShardsCoordinatorMock(5),
		ScrForwarder:    &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:  &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler:    &mock.FeeAccumulatorStub{},
		TxLogsProcessor: &mock.TxLogsProcessorStub{},
		EconomicsFee: &mock.FeeHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.0
			},
		},
		TxTypeHandler: &mock.TxTypeHandlerMock{},
		GasHandler: &mock.GasHandlerMock{
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		BuiltInFunctions: builtInFunctions.NewBuiltInFunctionContainer(),
	}
}

func TestNewSmartContractProcessorNilVM(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNoVM, err)
}

func TestNewSmartContractProcessorNilArgsParser(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ArgsParser = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilArgumentParser, err)
}

func TestNewSmartContractProcessorNilHasher(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.Hasher = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilHasher, err)
}

func TestNewSmartContractProcessorNilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.Marshalizer = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewSmartContractProcessorNilAccountsDB(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestNewSmartContractProcessorNilAdrConv(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.PubkeyConv = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewSmartContractProcessorNilShardCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.Coordinator = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewSmartContractProcessorNilFakeAccountsHandler(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.BlockChainHook = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilTemporaryAccountsHandler, err)
}

func TestNewSmartContractProcessor_NilIntermediateMock(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.ScrForwarder = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilIntermediateTransactionHandler, err)
}

func TestNewSmartContractProcessor_ErrNilUnsignedTxHandlerMock(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.TxFeeHandler = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilUnsignedTxHandler, err)
}

func TestNewSmartContractProcessor_ErrErrNilGasHandlerMock(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	arguments.GasHandler = nil
	sc, err := NewSmartContractProcessor(arguments)

	require.Nil(t, sc)
	require.Equal(t, process.ErrNilGasHandler, err)
}

func TestNewSmartContractProcessor(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)
	require.False(t, sc.IsInterfaceNil())
}

func TestScProcessor_DeploySmartContractBadParse(t *testing.T) {
	t.Parallel()

	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = &mock.VMContainerMock{}
	arguments.ArgsParser = argParser

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	parseError := fmt.Errorf("fooError")
	argParser.ParseDeployDataCalled = func(data string) (*parsers.DeployArgs, error) {
		return nil, parseError
	}

	arguments.AccountsDB = &mock.AccountsStub{
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
		LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return acntSrc, nil
		},
	}

	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	returnCode, err := sc.DeploySmartContract(tx, acntSrc)
	require.Equal(t, parseError, GetLatestTestError(sc))
	require.Equal(t, vmcommon.UserError, returnCode)
	require.Equal(t, uint64(1), acntSrc.GetNonce())
	require.True(t, acntSrc.GetBalance().Cmp(tx.Value) == 0)
}

func TestScProcessor_DeploySmartContractRunError(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &mock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}

	createError := fmt.Errorf("fooError")
	vm.RunSmartContractCreateCalled = func(input *vmcommon.ContractCreateInput) (output *vmcommon.VMOutput, e error) {
		return nil, createError
	}

	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	_, _ = sc.DeploySmartContract(tx, acntSrc)
	require.Equal(t, createError, GetLatestTestError(sc))
}

func TestScProcessor_DeploySmartContractDisabled(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &mock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	arguments.DisableDeploy = true
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	_, _ = sc.DeploySmartContract(tx, acntSrc)
	require.Equal(t, process.ErrSmartContractDeploymentIsDisabled, GetLatestTestError(sc))
}

func TestScProcessor_BuiltInCallSmartContractDisabled(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &mock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	arguments.DisableBuiltIn = true
	funcName := "builtIn"
	_ = arguments.BuiltInFunctions.Add(funcName, &mock.BuiltInFunctionStub{})
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte(funcName + "@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, nil)
	require.Equal(t, process.ErrFailedTransaction, err)
}

func TestScProcessor_BuiltInCallSmartContractSenderFailed(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = &mock.AccountsStub{RevertToSnapshotCalled: func(snapshot int) error {
		return nil
	}}
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	arguments.DisableBuiltIn = true
	funcName := "builtIn"
	localError := errors.New("failed built in call")
	_ = arguments.BuiltInFunctions.Add(funcName, &mock.BuiltInFunctionStub{
		ProcessBuiltinFunctionCalled: func(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			return nil, localError
		},
	})

	scrAdded := false
	badTxAdded := false
	arguments.BadTxForwarder = &mock.IntermediateTransactionHandlerMock{
		AddIntermediateTransactionsCalled: func(txs []data.TransactionHandler) error {
			badTxAdded = true
			return nil
		},
	}
	arguments.ScrForwarder = &mock.IntermediateTransactionHandlerMock{
		AddIntermediateTransactionsCalled: func(txs []data.TransactionHandler) error {
			scrAdded = true
			return nil
		},
	}

	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte(funcName + "@0500@0000")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	vm := &mock.VMExecutionHandlerStub{}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, nil)
	require.Equal(t, process.ErrFailedTransaction, err)
	require.True(t, scrAdded)
	require.True(t, badTxAdded)

	_, err = sc.ExecuteSmartContractTransaction(tx, nil, acntSrc)
	require.Nil(t, err)
}

func TestScProcessor_DeploySmartContractWrongTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	_, err = sc.DeploySmartContract(tx, acntSrc)
	require.Equal(t, process.ErrWrongTransaction, err)
}

func TestScProcessor_DeploySmartContract(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = generateEmptyByteSlice(createMockPubkeyConverter().Len())
	tx.Data = []byte("abba@0500@0000")
	tx.Value = big.NewInt(0)
	acntSrc, _ := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acntSrc, nil
	}

	_, err = sc.DeploySmartContract(tx, acntSrc)
	require.Nil(t, err)
	require.Nil(t, GetLatestTestError(sc))
}

func TestScProcessor_ExecuteSmartContractTransactionNilTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	_, err = sc.ExecuteSmartContractTransaction(nil, acntSrc, acntDst)
	require.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_ExecuteSmartContractTransactionNilAccount(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, _ := createAccounts(tx)

	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, nil)
	require.Equal(t, process.ErrNilSCDestAccount, err)

	acntSrc, acntDst := createAccounts(tx)
	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.Nil(t, err)

	acntSrc, _ = createAccounts(tx)
	acntDst = nil
	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.Equal(t, process.ErrNilSCDestAccount, err)
}

func TestScProcessor_ExecuteSmartContractTransactionBadParser(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	acntDst.SetCode([]byte("code"))
	tmpError := errors.New("error")
	called := false
	argParser.ParseCallDataCalled = func(data string) (string, [][]byte, error) {
		called = true
		return "", nil, tmpError
	}
	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.True(t, called)
	require.Nil(t, err)
}

func TestScProcessor_ExecuteSmartContractTransactionVMRunError(t *testing.T) {
	t.Parallel()

	vmContainer := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vmContainer
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST0000000")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	acntSrc, acntDst := createAccounts(tx)

	acntDst.SetCode([]byte("code"))
	tmpError := errors.New("error")
	vm := &mock.VMExecutionHandlerStub{}
	called := false
	vm.RunSmartContractCallCalled = func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
		called = true
		return nil, tmpError
	}
	vmContainer.GetCalled = func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
		return vm, nil
	}

	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.True(t, called)
	require.Nil(t, err)
}

func TestScProcessor_ExecuteSmartContractTransaction(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST0000000")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(0)
	acntSrc, acntDst := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acntSrc, nil
	}

	acntDst.SetCode([]byte("code"))
	_, err = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.Nil(t, err)
}

func TestScProcessor_ExecuteSmartContractTransactionSaveLogCalled(t *testing.T) {
	t.Parallel()

	slCalled := false

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	arguments.TxLogsProcessor = &mock.TxLogsProcessorStub{
		SaveLogCalled: func(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error {
			slCalled = true
			return nil
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST0000000")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(0)
	acntSrc, acntDst := createAccounts(tx)

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acntSrc, nil
	}

	acntDst.SetCode([]byte("code"))
	_, _ = sc.ExecuteSmartContractTransaction(tx, acntSrc, acntDst)
	require.True(t, slCalled)
}

func TestScProcessor_CreateVMCallInputWrongCode(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	tmpError := errors.New("error")
	argParser.ParseCallDataCalled = func(data string) (string, [][]byte, error) {
		return "", nil, tmpError
	}
	input, err := sc.createVMCallInput(tx)
	require.Nil(t, input)
	require.Equal(t, tmpError, err)
}

func TestScProcessor_CreateVMCallInput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	input, err := sc.createVMCallInput(tx)
	require.NotNil(t, input)
	require.Nil(t, err)
}

func TestScProcessor_CreateVMDeployBadCode(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Data = nil
	tx.Value = big.NewInt(0)

	badCodeError := errors.New("fooError")
	argParser.ParseDeployDataCalled = func(data string) (*parsers.DeployArgs, error) {
		return nil, badCodeError
	}

	input, vmType, err := sc.createVMDeployInput(tx)
	require.Nil(t, vmType)
	require.Nil(t, input)
	require.Equal(t, badCodeError, err)
}

func TestScProcessor_CreateVMDeployInput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("foobar")
	tx.Value = big.NewInt(45)

	expectedVMType := []byte{5, 6}
	expectedCodeMetadata := vmcommon.CodeMetadata{Upgradeable: true}
	argParser.ParseDeployDataCalled = func(data string) (*parsers.DeployArgs, error) {
		return &parsers.DeployArgs{
			Code:         []byte("code"),
			VMType:       expectedVMType,
			CodeMetadata: expectedCodeMetadata,
			Arguments:    nil,
		}, nil
	}

	input, vmType, err := sc.createVMDeployInput(tx)
	require.NotNil(t, input)
	require.Nil(t, err)
	require.Equal(t, vmcommon.DirectCall, input.CallType)
	require.True(t, bytes.Equal(expectedVMType, vmType))
	require.Equal(t, expectedCodeMetadata.ToBytes(), input.ContractCodeMetadata)
	require.Nil(t, err)
}

func TestScProcessor_CreateVMDeployInputNotEnoughArguments(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := NewArgumentParser()
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data@0000")
	tx.Value = big.NewInt(45)

	input, vmType, err := sc.createVMDeployInput(tx)
	require.Nil(t, input)
	require.Nil(t, vmType)
	require.Equal(t, parsers.ErrInvalidDeployArguments, err)
}

func TestScProcessor_CreateVMDeployInputWrongArgument(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	tmpError := errors.New("fooError")
	argParser.ParseDeployDataCalled = func(data string) (*parsers.DeployArgs, error) {
		return nil, tmpError
	}
	input, vmType, err := sc.createVMDeployInput(tx)
	require.Nil(t, input)
	require.Nil(t, vmType)
	require.Equal(t, tmpError, err)
}

func TestScProcessor_InitializeVMInputFromTx_ShouldErrNotEnoughGas(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.EconomicsFee = &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return 1000
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)
	tx.GasLimit = 100

	vmInput := &vmcommon.VMInput{}
	err = sc.initializeVMInputFromTx(vmInput, tx)
	require.Equal(t, process.ErrNotEnoughGas, err)
}

func TestScProcessor_InitializeVMInputFromTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	vmInput := &vmcommon.VMInput{}
	err = sc.initializeVMInputFromTx(vmInput, tx)
	require.Nil(t, err)
}

func createAccountsAndTransaction() (state.UserAccountHandler, state.UserAccountHandler, *transaction.Transaction) {
	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	acntSrc, acntDst := createAccounts(tx)

	return acntSrc.(state.UserAccountHandler), acntDst.(state.UserAccountHandler), tx
}

func TestScProcessor_processVMOutputNilVMOutput(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acntSrc, _, tx := createAccountsAndTransaction()

	txHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, tx)
	_, _, err = sc.processVMOutput(nil, txHash, tx, acntSrc, vmcommon.DirectCall)
	require.Equal(t, process.ErrNilVMOutput, err)
}

func TestScProcessor_processVMOutputNilTx(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acntSrc, _, _ := createAccountsAndTransaction()

	vmOutput := &vmcommon.VMOutput{}
	_, _, err = sc.processVMOutput(vmOutput, nil, nil, acntSrc, vmcommon.DirectCall)
	require.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_processVMOutputNilSndAcc(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}

	vmOutput := &vmcommon.VMOutput{
		GasRefund:    big.NewInt(0),
		GasRemaining: 0,
	}
	txHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, tx)
	_, _, err = sc.processVMOutput(vmOutput, txHash, tx, nil, vmcommon.DirectCall)
	require.Nil(t, err)
}

func TestScProcessor_processVMOutputNilDstAcc(t *testing.T) {
	t.Parallel()

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acntSnd, _, tx := createAccountsAndTransaction()

	vmOutput := &vmcommon.VMOutput{
		GasRefund:    big.NewInt(0),
		GasRemaining: 0,
	}

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acntSnd, nil
	}

	tx.Value = big.NewInt(0)
	txHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, tx)
	_, _, err = sc.processVMOutput(vmOutput, txHash, tx, acntSnd, vmcommon.DirectCall)
	require.Nil(t, err)
}

func TestScProcessor_GetAccountFromAddressAccNotFound(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return nil, state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("SRC"))
	require.Nil(t, acc)
	require.Equal(t, state.ErrAccNotFound, err)
}

func TestScProcessor_GetAccountFromAddrFailedGetExistingAccount(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		getCalled++
		return nil, state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("DST"))
	require.Nil(t, acc)
	require.Equal(t, state.ErrAccNotFound, err)
	require.Equal(t, 1, getCalled)
}

func TestScProcessor_GetAccountFromAddrAccNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		getCalled++
		return nil, state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId() + 1
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("DST"))
	require.Nil(t, acc)
	require.Nil(t, err)
	require.Equal(t, 0, getCalled)
}

func TestScProcessor_GetAccountFromAddr(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	getCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		getCalled++
		acc, _ := state.NewUserAccount(address)
		return acc, nil
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	acc, err := sc.getAccountFromAddress([]byte("DST"))
	require.NotNil(t, acc)
	require.Nil(t, err)
	require.Equal(t, 1, getCalled)
}

func TestScProcessor_DeleteAccountsFailedAtRemove(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return nil, state.ErrAccNotFound
	}
	accountsDB.RemoveAccountCalled = func(address []byte) error {
		removeCalled++
		return state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	deletedAccounts := make([][]byte, 0)
	deletedAccounts = append(deletedAccounts, []byte("acc1"), []byte("acc2"), []byte("acc3"))
	err = sc.deleteAccounts(deletedAccounts)
	require.Equal(t, state.ErrAccNotFound, err)
	require.Equal(t, 0, removeCalled)
}

func TestScProcessor_DeleteAccountsNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.RemoveAccountCalled = func(address []byte) error {
		removeCalled++
		return state.ErrAccNotFound
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	computeIdCalled := 0
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		computeIdCalled++
		return shardCoordinator.SelfId() + 1
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	deletedAccounts := make([][]byte, 0)
	deletedAccounts = append(deletedAccounts, []byte("acc1"), []byte("acc2"), []byte("acc3"))
	err = sc.deleteAccounts(deletedAccounts)
	require.Nil(t, err)
	require.Equal(t, 0, removeCalled)
	require.Equal(t, len(deletedAccounts), computeIdCalled)
}

func TestScProcessor_DeleteAccountsInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	removeCalled := 0
	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		acc, _ := state.NewUserAccount(address)
		return acc, nil
	}
	accountsDB.RemoveAccountCalled = func(address []byte) error {
		removeCalled++
		return nil
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	computeIdCalled := 0
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		computeIdCalled++
		return shardCoordinator.SelfId()
	}

	vm := &mock.VMContainerMock{}
	argParser := &mock.ArgumentParserMock{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.VmContainer = vm
	arguments.ArgsParser = argParser
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	deletedAccounts := make([][]byte, 0)
	deletedAccounts = append(deletedAccounts, []byte("acc1"), []byte("acc2"), []byte("acc3"))
	err = sc.deleteAccounts(deletedAccounts)
	require.Nil(t, err)
	require.Equal(t, len(deletedAccounts), removeCalled)
	require.Equal(t, len(deletedAccounts), computeIdCalled)
}

func TestScProcessor_ProcessSCPaymentAccNotInShardShouldNotReturnError(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10

	err = sc.processSCPayment(tx, nil)
	require.Nil(t, err)
}

func TestScProcessor_ProcessSCPaymentNotEnoughBalance(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 15

	acntSrc, _ := state.NewUserAccount(tx.SndAddr)
	_ = acntSrc.AddToBalance(big.NewInt(45))

	currBalance := acntSrc.GetBalance().Uint64()

	err = sc.processSCPayment(tx, acntSrc)
	require.Equal(t, process.ErrInsufficientFunds, err)
	require.Equal(t, currBalance, acntSrc.GetBalance().Uint64())
}

func TestScProcessor_ProcessSCPayment(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10

	acntSrc, _ := createAccounts(tx)
	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()
	modifiedBalance := currBalance - tx.Value.Uint64() - tx.GasLimit*tx.GasLimit

	err = sc.processSCPayment(tx, acntSrc)
	require.Nil(t, err)
	require.Equal(t, modifiedBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
}

func TestScProcessor_RefundGasToSenderNilAndZeroRefund(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10

	txHash := []byte("txHash")

	acntSrc, _ := createAccounts(tx)
	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(0)}
	_, _ = sc.createSCRForSender(
		vmOutput.GasRefund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		"",
		tx,
		txHash,
		acntSrc,
		vmcommon.DirectCall,
	)
	require.Nil(t, err)
	require.Equal(t, currBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
}

func TestScProcessor_RefundGasToSenderAccNotInShard(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)

	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 10
	txHash := []byte("txHash")
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(10)}
	sctx, consumed := sc.createSCRForSender(
		vmOutput.GasRefund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		"",
		tx,
		txHash,
		nil,
		vmcommon.DirectCall,
	)
	require.Nil(t, err)
	require.NotNil(t, sctx)
	require.Equal(t, 0, consumed.Cmp(big.NewInt(0).SetUint64(tx.GasPrice*tx.GasLimit)))

	vmOutput = &vmcommon.VMOutput{GasRemaining: 0, GasRefund: big.NewInt(10)}
	sctx, consumed = sc.createSCRForSender(
		vmOutput.GasRefund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		"",
		tx,
		txHash,
		nil,
		vmcommon.DirectCall,
	)
	require.Nil(t, err)
	require.NotNil(t, sctx)
	require.Equal(t, 0, consumed.Cmp(big.NewInt(0).SetUint64(tx.GasPrice*tx.GasLimit)))
}

func TestScProcessor_RefundGasToSender(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(10)
	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = &mock.FeeHandlerStub{MinGasPriceCalled: func() uint64 {
		return minGasPrice
	}}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 15
	tx.GasLimit = 15
	txHash := []byte("txHash")
	acntSrc, _ := createAccounts(tx)
	currBalance := acntSrc.(state.UserAccountHandler).GetBalance().Uint64()

	refundGas := big.NewInt(10)
	vmOutput := &vmcommon.VMOutput{GasRemaining: 0, GasRefund: refundGas}
	scr, _ := sc.createSCRForSender(
		vmOutput.GasRefund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		"",
		tx,
		txHash,
		acntSrc,
		vmcommon.DirectCall,
	)
	require.Nil(t, err)

	finalValue := big.NewInt(0).Mul(refundGas, big.NewInt(0).SetUint64(sc.economicsFee.MinGasPrice()))
	require.Equal(t, currBalance, acntSrc.(state.UserAccountHandler).GetBalance().Uint64())
	require.Equal(t, scr.Value.Cmp(finalValue), 0)
}

func TestScProcessor_DoNotRefundGasToSenderForAsyncCall(t *testing.T) {
	t.Parallel()

	minGasPrice := uint64(10)
	arguments := createMockSmartContractProcessorArguments()
	arguments.EconomicsFee = &mock.FeeHandlerStub{MinGasPriceCalled: func() uint64 {
		return minGasPrice
	}}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &smartContractResult.SmartContractResult{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 15
	tx.GasLimit = 15
	txHash := []byte("txHash")
	acntSrc, _ := state.NewUserAccount(tx.SndAddr)

	refundGas := big.NewInt(10)
	vmOutput := &vmcommon.VMOutput{GasRemaining: 10, GasRefund: refundGas}
	scr, _ := sc.createSCRForSender(
		vmOutput.GasRefund,
		vmOutput.GasRemaining,
		vmOutput.ReturnCode,
		vmOutput.ReturnData,
		"",
		tx,
		txHash,
		acntSrc,
		vmcommon.AsynchronousCall,
	)
	require.Nil(t, err)

	scrValue := big.NewInt(0).Mul(vmOutput.GasRefund, big.NewInt(0).SetUint64(sc.economicsFee.MinGasPrice()))
	require.Equal(t, scr.Value.Cmp(scrValue), 0)
	require.Equal(t, scr.GasLimit, vmOutput.GasRemaining)
	require.Equal(t, scr.GasPrice, tx.GasPrice)
}

func TestScProcessor_processVMOutputNilOutput(t *testing.T) {
	t.Parallel()

	acntSrc, _, tx := createAccountsAndTransaction()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)
	txHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, tx)
	_, _, err = sc.processVMOutput(nil, txHash, tx, acntSrc, vmcommon.DirectCall)

	require.Equal(t, process.ErrNilVMOutput, err)
}

func TestScProcessor_processVMOutputNilTransaction(t *testing.T) {
	t.Parallel()

	acntSrc, _, _ := createAccountsAndTransaction()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	vmOutput := &vmcommon.VMOutput{}
	_, _, err = sc.processVMOutput(vmOutput, nil, nil, acntSrc, vmcommon.DirectCall)

	require.Equal(t, process.ErrNilTransaction, err)
}

func TestScProcessor_processVMOutput(t *testing.T) {
	t.Parallel()

	acntSrc, _, tx := createAccountsAndTransaction()

	accntState := &mock.AccountsStub{}
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accntState
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	vmOutput := &vmcommon.VMOutput{
		GasRefund:    big.NewInt(0),
		GasRemaining: 0,
	}

	accntState.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		return acntSrc, nil
	}

	tx.Value = big.NewInt(0)
	txHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, tx)
	_, _, err = sc.processVMOutput(vmOutput, txHash, tx, acntSrc, vmcommon.DirectCall)
	require.Nil(t, err)
}

func TestScProcessor_processSCOutputAccounts(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}

	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}
	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	_, err = sc.processSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Code = []byte("contract-code")
	outacc1.Nonce = 5
	outacc1.BalanceDelta = big.NewInt(int64(5))
	outputAccounts = append(outputAccounts, outacc1)

	testAddr := outaddress
	testAcc, _ := state.NewUserAccount(testAddr)

	accountsDB.LoadAccountCalled = func(address []byte) (handler state.AccountHandler, e error) {
		if bytes.Equal(address, testAddr) {
			return testAcc, nil
		}
		return nil, state.ErrAccNotFound
	}

	accountsDB.SaveAccountCalled = func(accountHandler state.AccountHandler) error {
		return nil
	}

	tx.Value = big.NewInt(int64(5))
	_, err = sc.processSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outacc1.BalanceDelta = nil
	outacc1.Nonce++
	tx.Value = big.NewInt(0)
	_, err = sc.processSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outacc1.Nonce++
	outacc1.BalanceDelta = big.NewInt(int64(10))
	tx.Value = big.NewInt(int64(10))

	currentBalance := testAcc.Balance.Uint64()
	vmOutBalance := outacc1.BalanceDelta.Uint64()
	_, err = sc.processSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)
	require.Equal(t, currentBalance+vmOutBalance, testAcc.Balance.Uint64())
}

func TestScProcessor_processSCOutputAccountsNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	tx := &transaction.Transaction{Value: big.NewInt(0)}
	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	_, err = sc.processSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)

	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Code = []byte("contract-code")
	outacc1.Nonce = 5
	outputAccounts = append(outputAccounts, outacc1)

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.SelfId() + 1
	}

	_, err = sc.processSCOutputAccounts(outputAccounts, tx, []byte("hash"))
	require.Nil(t, err)
}

func TestScProcessor_CreateCrossShardTransactions(t *testing.T) {
	t.Parallel()

	testAccounts, _ := state.NewUserAccount([]byte("address"))
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, err error) {
			return testAccounts, nil
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	outputAccounts := make([]*vmcommon.OutputAccount, 0)
	outaddress := []byte("newsmartcontract")
	outacc1 := &vmcommon.OutputAccount{}
	outacc1.Address = outaddress
	outacc1.Nonce = 0
	outacc1.Balance = big.NewInt(int64(5))
	outacc1.BalanceDelta = big.NewInt(int64(15))
	outputAccounts = append(outputAccounts, outacc1, outacc1, outacc1)

	tx := &transaction.Transaction{}
	tx.Nonce = 1
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")

	tx.Value = big.NewInt(45)
	tx.GasPrice = 10
	tx.GasLimit = 15
	txHash := []byte("txHash")

	scTxs, err := sc.processSCOutputAccounts(outputAccounts, tx, txHash)
	require.Nil(t, err)
	require.Equal(t, len(outputAccounts), len(scTxs))
}

func TestScProcessor_ProcessSmartContractResultNilScr(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	_, err = sc.ProcessSmartContractResult(nil)
	require.Equal(t, process.ErrNilSmartContractResult, err)
}

func TestScProcessor_ProcessSmartContractResultErrGetAccount(t *testing.T) {
	t.Parallel()

	accError := errors.New("account get error")
	called := false
	accountsDB := &mock.AccountsStub{LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
		called = true
		return nil, accError
	}}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{RcvAddr: []byte("recv address")}
	_, _ = sc.ProcessSmartContractResult(&scr)
	require.True(t, called)
}

func TestScProcessor_ProcessSmartContractResultAccNotInShard(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.CurrentShard + 1
	}
	scr := smartContractResult.SmartContractResult{RcvAddr: []byte("recv address")}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Equal(t, err, process.ErrNilSCDestAccount)
}

func TestScProcessor_ProcessSmartContractResultBadAccType(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return &mock.AccountWrapMock{}, nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{RcvAddr: []byte("recv address"), Value: big.NewInt(0)}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
}

func TestScProcessor_ProcessSmartContractResultNotPayable(t *testing.T) {
	t.Parallel()

	userAcc, _ := state.NewUserAccount([]byte("recv address"))
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			if bytes.Equal(address, userAcc.Address) {
				return userAcc, nil
			}
			return state.NewEmptyUserAccount(), nil
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.BlockChainHook = &mock.BlockChainHookHandlerMock{
		IsPayableCalled: func(address []byte) (bool, error) {
			return false, nil
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		RcvAddr: userAcc.Address,
		SndAddr: []byte("snd addr"),
		Value:   big.NewInt(0),
	}
	returnCode, err := sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)

	scr.Value = big.NewInt(10)
	returnCode, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, returnCode)
	require.True(t, userAcc.GetBalance().Cmp(zero) == 0)

	scr.OriginalSender = scr.RcvAddr
	returnCode, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	require.True(t, userAcc.GetBalance().Cmp(scr.Value) == 0)
}

func TestScProcessor_ProcessSmartContractResultOutputBalanceNil(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return state.NewUserAccount(address)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		RcvAddr: []byte("recv address"),
		SndAddr: []byte("snd addr"),
		Value:   big.NewInt(0),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
}

func TestScProcessor_ProcessSmartContractResultWithCode(t *testing.T) {
	t.Parallel()

	putCodeCalled := 0
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return state.NewUserAccount(address)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			putCodeCalled++
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		SndAddr: []byte("snd addr"),
		RcvAddr: []byte("recv address"),
		Code:    []byte("code"),
		Value:   big.NewInt(15),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.Equal(t, 1, putCodeCalled)
}

func TestScProcessor_ProcessSmartContractResultWithData(t *testing.T) {
	t.Parallel()

	saveAccountCalled := 0
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return state.NewUserAccount(address)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			saveAccountCalled++
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	test := "test"
	result := ""
	sep := "@"
	for i := 0; i < 6; i++ {
		result += test
		result += sep
	}

	scr := smartContractResult.SmartContractResult{
		SndAddr: []byte("snd addr"),
		RcvAddr: []byte("recv address"),
		Data:    []byte(result),
		Value:   big.NewInt(15),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.Equal(t, 1, saveAccountCalled)
}

func TestScProcessor_ProcessSmartContractResultDeploySCShouldError(t *testing.T) {
	t.Parallel()

	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			return state.NewUserAccount(address)
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.TxTypeHandler = &mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType) {
			return process.SCDeployment
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		SndAddr: []byte("snd addr"),
		RcvAddr: []byte("recv address"),
		Data:    []byte("code@06"),
		Value:   big.NewInt(15),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
}

func TestScProcessor_ProcessSmartContractResultExecuteSC(t *testing.T) {
	t.Parallel()

	scAddress := []byte("000000000001234567890123456789012")
	dstScAddress, _ := state.NewUserAccount(scAddress)
	dstScAddress.SetCode([]byte("code"))
	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			if bytes.Equal(scAddress, address) {
				return dstScAddress, nil
			}
			return nil, nil
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(scAddress, address) {
			return shardCoordinator.SelfId()
		}
		return shardCoordinator.SelfId() + 1
	}

	executeCalled := false
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return &mock.VMExecutionHandlerStub{
				RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
					executeCalled = true
					return nil, nil
				},
			}, nil
		},
	}
	arguments.TxTypeHandler = &mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType) {
			return process.SCInvoking
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		SndAddr: []byte("snd addr"),
		RcvAddr: scAddress,
		Data:    []byte("code@06"),
		Value:   big.NewInt(15),
	}
	_, err = sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.True(t, executeCalled)
}

func TestScProcessor_ProcessRelayedSCRValueBackToRelayer(t *testing.T) {
	t.Parallel()

	scAddress := []byte("000000000001234567890123456789012")
	dstScAddress, _ := state.NewUserAccount(scAddress)
	dstScAddress.SetCode([]byte("code"))

	baseValue := big.NewInt(100)
	userAddress := []byte("111111111111234567890123456789012")
	userAcc, _ := state.NewUserAccount(userAddress)
	_ = userAcc.AddToBalance(baseValue)
	relayedAddress := []byte("211111111111234567890123456789012")
	relayedAcc, _ := state.NewUserAccount(relayedAddress)

	accountsDB := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
			if bytes.Equal(scAddress, address) {
				return dstScAddress, nil
			}
			if bytes.Equal(userAddress, address) {
				return userAcc, nil
			}
			if bytes.Equal(relayedAddress, address) {
				return relayedAcc, nil
			}

			return nil, nil
		},
		SaveAccountCalled: func(accountHandler state.AccountHandler) error {
			return nil
		},
		RevertToSnapshotCalled: func(snapshot int) error {
			return nil
		},
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	executeCalled := false
	arguments := createMockSmartContractProcessorArguments()
	arguments.AccountsDB = accountsDB
	arguments.Coordinator = shardCoordinator
	arguments.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return &mock.VMExecutionHandlerStub{
				RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
					executeCalled = true
					return &vmcommon.VMOutput{ReturnCode: vmcommon.UserError}, nil
				},
			}, nil
		},
	}
	arguments.TxTypeHandler = &mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType) {
			return process.SCInvoking
		},
	}
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	scr := smartContractResult.SmartContractResult{
		SndAddr:      userAddress,
		RcvAddr:      scAddress,
		RelayerAddr:  relayedAddress,
		RelayedValue: big.NewInt(10),
		Data:         []byte("code@06"),
		Value:        big.NewInt(15),
	}
	returnCode, err := sc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)
	require.True(t, executeCalled)
	require.Equal(t, returnCode, vmcommon.UserError)

	require.True(t, relayedAcc.GetBalance().Cmp(scr.RelayedValue) == 0)
	userReturnValue := big.NewInt(0).Sub(scr.Value, scr.RelayedValue)
	userFinalValue := baseValue.Sub(baseValue, scr.Value)
	userFinalValue.Add(userFinalValue, userReturnValue)
	require.True(t, userAcc.GetBalance().Cmp(userFinalValue) == 0)
}

func TestScProcessor_checkUpgradePermission(t *testing.T) {
	t.Parallel()

	arguments := createMockSmartContractProcessorArguments()
	sc, err := NewSmartContractProcessor(arguments)
	require.NotNil(t, sc)
	require.Nil(t, err)

	// Not an upgrade
	err = sc.checkUpgradePermission(nil, &vmcommon.ContractCallInput{})
	require.Nil(t, err)

	// Upgrade, nil contract passed (quite impossible though)
	err = sc.checkUpgradePermission(nil, &vmcommon.ContractCallInput{Function: "upgradeContract"})
	require.Equal(t, process.ErrUpgradeNotAllowed, err)

	// Create a contract, owned by Alice
	contract, err := state.NewUserAccount([]byte("contract"))
	require.Nil(t, err)
	contract.SetOwnerAddress([]byte("alice"))
	// Not yet upgradeable
	contract.SetCodeMetadata([]byte{0, 0})

	// Upgrade as Alice, but not allowed (not upgradeable)
	err = sc.checkUpgradePermission(contract, &vmcommon.ContractCallInput{Function: "upgradeContract", VMInput: vmcommon.VMInput{CallerAddr: []byte("alice")}})
	require.Equal(t, process.ErrUpgradeNotAllowed, err)

	// Upgrade as Bob, but not allowed
	err = sc.checkUpgradePermission(contract, &vmcommon.ContractCallInput{Function: "upgradeContract", VMInput: vmcommon.VMInput{CallerAddr: []byte("bob")}})
	require.Equal(t, process.ErrUpgradeNotAllowed, err)

	// Mark as upgradeable
	contract.SetCodeMetadata([]byte{1, 0})

	// Upgrade as Alice, allowed
	err = sc.checkUpgradePermission(contract, &vmcommon.ContractCallInput{Function: "upgradeContract", VMInput: vmcommon.VMInput{CallerAddr: []byte("alice")}})
	require.Nil(t, err)

	// Upgrade as Bob, but not allowed
	err = sc.checkUpgradePermission(contract, &vmcommon.ContractCallInput{Function: "upgradeContract", VMInput: vmcommon.VMInput{CallerAddr: []byte("bob")}})
	require.Equal(t, process.ErrUpgradeNotAllowed, err)

	// Upgrade as nobody, not allowed
	err = sc.checkUpgradePermission(contract, &vmcommon.ContractCallInput{Function: "upgradeContract", VMInput: vmcommon.VMInput{CallerAddr: nil}})
	require.Equal(t, process.ErrUpgradeNotAllowed, err)
}
