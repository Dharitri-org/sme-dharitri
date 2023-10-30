package process

import (
	"github.com/Dharitri-org/sme-dharitri/node/external"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/vm"
)

type genesisProcessors struct {
	txCoordinator  process.TransactionCoordinator
	systemSCs      vm.SystemSCContainer
	txProcessor    process.TransactionProcessor
	scProcessor    process.SmartContractProcessor
	scrProcessor   process.SmartContractResultProcessor
	rwdProcessor   process.RewardTransactionProcessor
	blockchainHook process.BlockChainHookHandler
	queryService   external.SCQueryService
	vmContainer    process.VirtualMachinesContainer
}
