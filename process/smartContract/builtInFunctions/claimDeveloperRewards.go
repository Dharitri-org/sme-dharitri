package builtInFunctions

import (
	"bytes"
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

var _ process.BuiltinFunction = (*claimDeveloperRewards)(nil)

type claimDeveloperRewards struct {
	gasCost uint64
}

// NewClaimDeveloperRewardsFunc returns a new developer rewards implementation
func NewClaimDeveloperRewardsFunc(gasCost uint64) *claimDeveloperRewards {
	return &claimDeveloperRewards{gasCost: gasCost}
}

// ProcessBuiltinFunction processes the protocol built-in smart contract function
func (c *claimDeveloperRewards) ProcessBuiltinFunction(
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	if vmInput == nil {
		return nil, process.ErrNilVmInput
	}
	if vmInput.CallValue.Cmp(zero) != 0 {
		return nil, process.ErrBuiltInFunctionCalledWithValue
	}
	if check.IfNil(acntDst) {
		// cross-shard call, in sender shard only the gas is taken out
		return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
	}

	if !bytes.Equal(vmInput.CallerAddr, acntDst.GetOwnerAddress()) {
		return nil, process.ErrOperationNotPermitted
	}
	if vmInput.GasProvided < c.gasCost {
		return nil, process.ErrNotEnoughGas
	}

	value, err := acntDst.ClaimDeveloperRewards(vmInput.CallerAddr)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{GasRemaining: vmInput.GasProvided - c.gasCost}
	outputAcc := &vmcommon.OutputAccount{
		Address:      vmInput.CallerAddr,
		BalanceDelta: big.NewInt(0).Set(value),
	}
	vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	vmOutput.OutputAccounts[string(outputAcc.Address)] = outputAcc

	if check.IfNil(acntSnd) {
		return vmOutput, nil
	}

	err = acntSnd.AddToBalance(value)
	if err != nil {
		return nil, err
	}

	if core.IsSmartContractAddress(vmInput.CallerAddr) {
		vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
	}

	return vmOutput, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (c *claimDeveloperRewards) IsInterfaceNil() bool {
	return c == nil
}
