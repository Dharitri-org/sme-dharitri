package mock

import (
	"github.com/Dharitri-org/sme-dharitri/data/state"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

type BuiltInFunctionStub struct {
	ProcessBuiltinFunctionCalled func(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error)
}

// ProcessBuiltinFunction -
func (b *BuiltInFunctionStub) ProcessBuiltinFunction(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if b.ProcessBuiltinFunctionCalled != nil {
		return b.ProcessBuiltinFunctionCalled(acntSnd, acntDst, vmInput)
	}
	return &vmcommon.VMOutput{}, nil
}

// IsInterfaceNil -
func (b *BuiltInFunctionStub) IsInterfaceNil() bool {
	return b == nil
}
