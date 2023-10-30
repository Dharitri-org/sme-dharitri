package builtInFunctions

import (
	"bytes"
	"fmt"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

var _ process.BuiltinFunction = (*changeOwnerAddress)(nil)

type changeOwnerAddress struct {
	gasCost uint64
}

// NewChangeOwnerAddressFunc create a new change owner built in function
func NewChangeOwnerAddressFunc(gasCost uint64) *changeOwnerAddress {
	return &changeOwnerAddress{gasCost: gasCost}
}

// ProcessBuiltinFunction processes simple protocol built-in function
func (c *changeOwnerAddress) ProcessBuiltinFunction(
	_, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	if vmInput == nil {
		return nil, process.ErrNilVmInput
	}
	if len(vmInput.Arguments) == 0 {
		return nil, process.ErrInvalidArguments
	}
	if vmInput.CallValue.Cmp(zero) != 0 {
		return nil, process.ErrBuiltInFunctionCalledWithValue
	}
	if len(vmInput.Arguments[0]) != len(vmInput.CallerAddr) {
		return nil, process.ErrInvalidAddressLength
	}
	if vmInput.GasProvided < c.gasCost {
		return nil, process.ErrNotEnoughGas
	}
	if check.IfNil(acntDst) {
		// cross-shard call, in sender shard only the gas is taken out
		return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
	}

	if !bytes.Equal(vmInput.CallerAddr, acntDst.GetOwnerAddress()) {
		return nil, fmt.Errorf("%w not the owner of the account", process.ErrOperationNotPermitted)
	}

	err := acntDst.ChangeOwnerAddress(vmInput.CallerAddr, vmInput.Arguments[0])
	if err != nil {
		return nil, err
	}

	return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (c *changeOwnerAddress) IsInterfaceNil() bool {
	return c == nil
}
