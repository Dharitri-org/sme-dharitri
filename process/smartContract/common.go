package smartContract

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

func findVMByTransaction(container process.VirtualMachinesContainer, tx data.TransactionHandler) (vmcommon.VMExecutionHandler, error) {
	scAddress := tx.GetRcvAddr()
	return findVMByScAddress(container, scAddress)
}

func findVMByScAddress(container process.VirtualMachinesContainer, scAddress []byte) (vmcommon.VMExecutionHandler, error) {
	vmType, err := parseVMTypeFromContractAddress(scAddress)
	if err != nil {
		return nil, err
	}

	vm, err := container.Get(vmType)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func parseVMTypeFromContractAddress(contractAddress []byte) ([]byte, error) {
	// TODO: Why not check against AddressLength (32)?
	if len(contractAddress) < core.NumInitCharactersForScAddress {
		return nil, process.ErrInvalidVMType
	}

	startIndex := core.NumInitCharactersForScAddress - core.VMTypeLen
	endIndex := core.NumInitCharactersForScAddress
	return contractAddress[startIndex:endIndex], nil
}
