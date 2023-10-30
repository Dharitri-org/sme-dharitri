package builtInFunctions

import (
	"math/big"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/mock"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestDCTTransfer_ProcessBuiltInFunctionErrors(t *testing.T) {
	t.Parallel()

	dct, _ := NewDCTTransferFunc(10, &mock.MarshalizerMock{})
	_, err := dct.ProcessBuiltinFunction(nil, nil, nil)
	assert.Equal(t, err, process.ErrNilVmInput)

	input := &vmcommon.ContractCallInput{}
	_, err = dct.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrInvalidArguments)

	input = &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := []byte("value")
	input.Arguments = [][]byte{key, value}
	_, err = dct.ProcessBuiltinFunction(nil, nil, input)
	assert.Nil(t, err)

	input.GasProvided = dct.funcGasCost - 1
	accSnd := state.NewEmptyUserAccount()
	_, err = dct.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Equal(t, err, process.ErrNotEnoughGas)
}

func TestDCTTransfer_ProcessBuiltInFunctionSingleShard(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	dct, _ := NewDCTTransferFunc(10, marshalizer)

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accSnd, _ := state.NewUserAccount([]byte("snd"))
	accDst, _ := state.NewUserAccount([]byte("dst"))

	_, err := dct.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Equal(t, err, process.ErrInsufficientFunds)

	dctKey := append(dct.keyPrefix, key...)
	dctToken := &DCToken{Value: big.NewInt(100)}
	marshalledData, _ := marshalizer.Marshal(dctToken)
	accSnd.DataTrieTracker().SaveKeyValue(dctKey, marshalledData)

	_, err = dct.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Nil(t, err)
	marshalledData, _ = accSnd.DataTrieTracker().RetrieveValue(dctKey)
	_ = marshalizer.Unmarshal(dctToken, marshalledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(90)) == 0)

	marshalledData, _ = accDst.DataTrieTracker().RetrieveValue(dctKey)
	_ = marshalizer.Unmarshal(dctToken, marshalledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(10)) == 0)
}

func TestDCTTransfer_ProcessBuiltInFunctionSenderInShard(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	dct, _ := NewDCTTransferFunc(10, marshalizer)

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accSnd, _ := state.NewUserAccount([]byte("snd"))

	dctKey := append(dct.keyPrefix, key...)
	dctToken := &DCToken{Value: big.NewInt(100)}
	marshalledData, _ := marshalizer.Marshal(dctToken)
	accSnd.DataTrieTracker().SaveKeyValue(dctKey, marshalledData)

	_, err := dct.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Nil(t, err)
	marshalledData, _ = accSnd.DataTrieTracker().RetrieveValue(dctKey)
	_ = marshalizer.Unmarshal(dctToken, marshalledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(90)) == 0)
}

func TestDCTTransfer_ProcessBuiltInFunctionDestInShard(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	dct, _ := NewDCTTransferFunc(10, marshalizer)

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accDst, _ := state.NewUserAccount([]byte("dst"))

	_, err := dct.ProcessBuiltinFunction(nil, accDst, input)
	assert.Nil(t, err)
	dctKey := append(dct.keyPrefix, key...)
	dctToken := &DCToken{}
	marshalledData, _ := accDst.DataTrieTracker().RetrieveValue(dctKey)
	_ = marshalizer.Unmarshal(dctToken, marshalledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(10)) == 0)
}
