package builtInFunctions

import (
	"math/big"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/data/state"

	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
	"github.com/stretchr/testify/require"
)

func TestChangeOwnerAddress_ProcessBuiltinFunction(t *testing.T) {
	t.Parallel()

	coa := changeOwnerAddress{}

	owner := []byte("send")
	addr := []byte("addr")

	acc, _ := state.NewUserAccount(addr)
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{CallerAddr: owner, CallValue: big.NewInt(0)},
	}

	_, err := coa.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Equal(t, process.ErrInvalidArguments, err)

	newAddr := []byte("0000")
	vmInput.Arguments = [][]byte{newAddr}
	_, err = coa.ProcessBuiltinFunction(nil, acc, nil)
	require.Equal(t, process.ErrNilVmInput, err)

	_, err = coa.ProcessBuiltinFunction(nil, nil, vmInput)
	require.Nil(t, err)

	acc.OwnerAddress = owner
	_, err = coa.ProcessBuiltinFunction(nil, acc, vmInput)
	require.Nil(t, err)
}
