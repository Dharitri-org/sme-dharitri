package shard

import (
	core "github.com/Dharitri-org/sme-core-vm-go/core"
	coreHost "github.com/Dharitri-org/sme-core-vm-go/core/host"
	ipcCommon "github.com/Dharitri-org/sme-core-vm-go/ipc/common"
	ipcMarshaling "github.com/Dharitri-org/sme-core-vm-go/ipc/marshaling"
	ipcNodePart "github.com/Dharitri-org/sme-core-vm-go/ipc/nodepart"
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/factory"
	"github.com/Dharitri-org/sme-dharitri/process/factory/containers"
	"github.com/Dharitri-org/sme-dharitri/process/smartContract/hooks"
	logger "github.com/Dharitri-org/sme-logger"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

var _ process.VirtualMachinesContainerFactory = (*vmContainerFactory)(nil)

var logVMContainerFactory = logger.GetOrCreate("vmContainerFactory")

type vmContainerFactory struct {
	config             config.VirtualMachineConfig
	blockChainHookImpl *hooks.BlockChainHookImpl
	cryptoHook         vmcommon.CryptoHook
	blockGasLimit      uint64
	gasSchedule        map[string]map[string]uint64
	builtinFunctions   vmcommon.FunctionNames
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(
	config config.VirtualMachineConfig,
	blockGasLimit uint64,
	gasSchedule map[string]map[string]uint64,
	argBlockChainHook hooks.ArgBlockChainHook,
) (*vmContainerFactory, error) {
	if gasSchedule == nil {
		return nil, process.ErrNilGasSchedule
	}

	blockChainHookImpl, err := hooks.NewBlockChainHookImpl(argBlockChainHook)
	if err != nil {
		return nil, err
	}

	cryptoHook := hooks.NewVMCryptoHook()
	builtinFunctions := blockChainHookImpl.GetBuiltinFunctionNames()

	return &vmContainerFactory{
		config:             config,
		blockChainHookImpl: blockChainHookImpl,
		cryptoHook:         cryptoHook,
		blockGasLimit:      blockGasLimit,
		gasSchedule:        gasSchedule,
		builtinFunctions:   builtinFunctions,
	}, nil
}

// Create sets up all the needed virtual machine returning a container of all the VMs
func (vmf *vmContainerFactory) Create() (process.VirtualMachinesContainer, error) {
	container := containers.NewVirtualMachinesContainer()

	currVm, err := vmf.createCoreVM()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.CoreVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (vmf *vmContainerFactory) createCoreVM() (vmcommon.VMExecutionHandler, error) {
	if vmf.config.OutOfProcessEnabled {
		return vmf.createOutOfProcessCoreVM()
	}

	return vmf.createInProcessCoreVM()
}

func (vmf *vmContainerFactory) createOutOfProcessCoreVM() (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Info("createOutOfProcessCoreVM", "config", vmf.config)

	outOfProcessConfig := vmf.config.OutOfProcessConfig
	logsMarshalizer := ipcMarshaling.ParseKind(outOfProcessConfig.LogsMarshalizer)
	messagesMarshalizer := ipcMarshaling.ParseKind(outOfProcessConfig.MessagesMarshalizer)
	maxLoopTime := outOfProcessConfig.MaxLoopTime

	logger.GetLogLevelPattern()

	coreVM, err := ipcNodePart.NewCoreDriver(
		vmf.blockChainHookImpl,
		ipcCommon.CoreArguments{
			VMHostParameters: core.VMHostParameters{
				VMType:                     factory.CoreVirtualMachine,
				BlockGasLimit:              vmf.blockGasLimit,
				GasSchedule:                vmf.gasSchedule,
				ProtocolBuiltinFunctions:   vmf.builtinFunctions,
				DharitriProtectedKeyPrefix: []byte(core.DharitriProtectedKeyPrefix),
			},
			LogsMarshalizer:     logsMarshalizer,
			MessagesMarshalizer: messagesMarshalizer,
		},
		ipcNodePart.Config{MaxLoopTime: maxLoopTime},
	)
	return coreVM, err
}

func (vmf *vmContainerFactory) createInProcessCoreVM() (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Info("createInProcessCoreVM")
	return coreHost.NewCoreVM(
		vmf.blockChainHookImpl,
		vmf.cryptoHook,
		&core.VMHostParameters{
			VMType:                     factory.CoreVirtualMachine,
			BlockGasLimit:              vmf.blockGasLimit,
			GasSchedule:                vmf.gasSchedule,
			ProtocolBuiltinFunctions:   vmf.builtinFunctions,
			DharitriProtectedKeyPrefix: []byte(core.DharitriProtectedKeyPrefix),
		},
	)
}

// BlockChainHookImpl returns the created blockChainHookImpl
func (vmf *vmContainerFactory) BlockChainHookImpl() process.BlockChainHookHandler {
	return vmf.blockChainHookImpl
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	return vmf == nil
}
