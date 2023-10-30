package systemSmartContracts

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/vm"
	"github.com/Dharitri-org/sme-dharitri/vm/mock"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
	"github.com/stretchr/testify/assert"
)

func createMockArgumentsForDCT() ArgsNewDCTSmartContract {
	return ArgsNewDCTSmartContract{
		Eei:     &mock.SystemEIStub{},
		GasCost: vm.GasCost{MetaChainSystemSCsCost: vm.MetaChainSystemSCsCost{DCTIssue: 10}},
		DCTSCConfig: config.DCTSystemSCConfig{
			BaseIssuingCost: "1000",
		},
		DCTSCAddress: []byte("address"),
		Marshalizer:  &mock.MarshalizerMock{},
		Hasher:       &mock.HasherMock{},
	}
}

func TestNewDCTSmartContract(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCT()
	e, err := NewDCTSmartContract(args)
	ky := hex.EncodeToString([]byte("DHARITRIdcttxgenDCTtkn"))
	fmt.Println(ky)

	assert.Nil(t, err)
	assert.NotNil(t, e)
}

func TestDct_ExecuteIssue(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCT()
	e, _ := NewDCTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("addr"),
			Arguments:      nil,
			CallValue:      big.NewInt(0),
			CallType:       0,
			GasPrice:       0,
			GasProvided:    0,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issue",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = [][]byte{[]byte("name"), []byte("1000")}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments[0] = []byte("01234567891")
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.DCTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.DCTIssue
	output = e.Execute(vmInput)

	assert.Equal(t, vmcommon.Ok, output)
}

func TestDct_ExecuteInit(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCT()
	e, _ := NewDCTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("addr"),
			Arguments:      nil,
			CallValue:      big.NewInt(0),
			CallType:       0,
			GasPrice:       0,
			GasProvided:    0,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      "_init",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.Ok, output)
}

func TestDct_ExecuteIssueProtected(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForDCT()
	e, _ := NewDCTSmartContract(args)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:     []byte("addr"),
			Arguments:      nil,
			CallValue:      big.NewInt(0),
			CallType:       0,
			GasPrice:       0,
			GasProvided:    0,
			OriginalTxHash: nil,
			CurrentTxHash:  nil,
		},
		RecipientAddr: []byte("addr"),
		Function:      "issueProtected",
	}
	output := e.Execute(vmInput)
	assert.Equal(t, vmcommon.UserError, output)

	vmInput.CallerAddr = e.ownerAddress
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = [][]byte{[]byte("name"), []byte("1000")}
	output = e.Execute(vmInput)
	assert.Equal(t, vmcommon.FunctionWrongSignature, output)

	vmInput.Arguments = [][]byte{[]byte("newOwner"), []byte("name"), []byte("1000")}

	vmInput.Arguments[0] = []byte("01234567891")
	vmInput.CallValue, _ = big.NewInt(0).SetString(args.DCTSCConfig.BaseIssuingCost, 10)
	vmInput.GasProvided = args.GasCost.MetaChainSystemSCsCost.DCTIssue
	output = e.Execute(vmInput)

	assert.Equal(t, vmcommon.Ok, output)
}
