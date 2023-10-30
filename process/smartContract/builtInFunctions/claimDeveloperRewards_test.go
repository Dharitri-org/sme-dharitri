package builtInFunctions

import (
	"math/big"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
	"github.com/stretchr/testify/require"
)

func TestClaimDeveloperRewards_ProcessBuiltinFunction(t *testing.T) {
	t.Parallel()

	cdr := claimDeveloperRewards{}

	sender := []byte("sender")
	acc, _ := state.NewUserAccount([]byte("addr12"))

	vmOutput, err := cdr.ProcessBuiltinFunction(nil, acc, nil)
	require.Nil(t, vmOutput)
	require.Equal(t, process.ErrNilVmInput, err)

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  sender,
			GasProvided: 100,
			CallValue:   big.NewInt(0),
		},
	}
	vmOutput, err = cdr.ProcessBuiltinFunction(nil, nil, vmInput)
	require.Nil(t, err)
	require.NotNil(t, vmOutput)

	vmOutput, err = cdr.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Nil(t, vmOutput)
	require.Equal(t, state.ErrOperationNotPermitted, err)

	acc.OwnerAddress = sender
	value := big.NewInt(100)
	acc.AddToDeveloperReward(value)
	vmOutput, err = cdr.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Nil(t, err)
	require.Equal(t, value, vmOutput.OutputAccounts[string(vmInput.CallerAddr)].BalanceDelta)
}
