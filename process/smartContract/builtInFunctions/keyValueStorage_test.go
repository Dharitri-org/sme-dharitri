package builtInFunctions

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
	"github.com/stretchr/testify/require"
)

func TestSaveKeyValue_ProcessBuiltinFunction(t *testing.T) {
	t.Parallel()

	funcGasCost := uint64(1)
	gasConfig := BaseOperationCost{
		StorePerByte:    1,
		ReleasePerByte:  1,
		DataCopyPerByte: 1,
		PersistPerByte:  1,
		CompilePerByte:  1,
	}

	skv, _ := NewSaveKeyValueStorageFunc(gasConfig, funcGasCost)

	addr := []byte("addr")
	acc, _ := state.NewUserAccount(addr)
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  addr,
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
		RecipientAddr: addr,
	}

	_, err := skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, process.ErrInvalidArguments, err)

	_, err = skv.ProcessBuiltinFunction(nil, acc, nil)
	require.Equal(t, process.ErrNilVmInput, err)

	key := []byte("key")
	value := []byte("value")
	vmInput.Arguments = [][]byte{key, value}

	_, err = skv.ProcessBuiltinFunction(nil, nil, vmInput)
	require.Equal(t, process.ErrNilSCDestAccount, err)

	_, err = skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Nil(t, err)
	retrievedValue, _ := acc.DataTrieTracker().RetrieveValue(key)
	require.True(t, bytes.Equal(retrievedValue, value))

	vmInput.CallerAddr = []byte("other")
	_, err = skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.True(t, errors.Is(err, process.ErrOperationNotPermitted))

	key = []byte(core.DharitriProtectedKeyPrefix + "is the king")
	value = []byte("value")
	vmInput.Arguments = [][]byte{key, value}

	_, err = skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.True(t, errors.Is(err, process.ErrOperationNotPermitted))
}

func TestSaveKeyValue_ProcessBuiltinFunctionMultipleKeys(t *testing.T) {
	t.Parallel()

	funcGasCost := uint64(1)
	gasConfig := BaseOperationCost{
		StorePerByte:    1,
		ReleasePerByte:  1,
		DataCopyPerByte: 1,
		PersistPerByte:  1,
		CompilePerByte:  1,
	}
	skv, _ := NewSaveKeyValueStorageFunc(gasConfig, funcGasCost)

	addr := []byte("addr")
	acc, _ := state.NewUserAccount(addr)
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  addr,
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
		RecipientAddr: addr,
	}

	key := []byte("key")
	value := []byte("value")
	vmInput.Arguments = [][]byte{key, value, key, value, key}

	_, err := skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, err, process.ErrInvalidArguments)

	key2 := []byte("key2")
	value2 := []byte("value2")
	vmInput.Arguments = [][]byte{key, value, key2, value2}

	_, err = skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Nil(t, err)
	retrievedValue, _ := acc.DataTrieTracker().RetrieveValue(key)
	require.True(t, bytes.Equal(retrievedValue, value))
	retrievedValue, _ = acc.DataTrieTracker().RetrieveValue(key2)
	require.True(t, bytes.Equal(retrievedValue, value2))

	vmInput.GasProvided = 1
	vmInput.Arguments = [][]byte{[]byte("key3"), []byte("value")}
	_, err = skv.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, err, process.ErrNotEnoughGas)
}
