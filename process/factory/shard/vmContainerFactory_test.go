package shard

import (
	"testing"

	coreConfig "github.com/Dharitri-org/sme-core-vm-go/config"
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/factory"
	"github.com/Dharitri-org/sme-dharitri/process/mock"
	"github.com/Dharitri-org/sme-dharitri/process/smartContract/builtInFunctions"
	"github.com/Dharitri-org/sme-dharitri/process/smartContract/hooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockVMAccountsArguments() hooks.ArgBlockChainHook {
	arguments := hooks.ArgBlockChainHook{
		Accounts: &mock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
				return &mock.AccountWrapMock{}, nil
			},
		},
		PubkeyConv:       mock.NewPubkeyConverterMock(32),
		StorageService:   &mock.ChainStorerMock{},
		BlockChain:       &mock.BlockChainMock{},
		ShardCoordinator: mock.NewOneShardCoordinatorMock(),
		Marshalizer:      &mock.MarshalizerMock{},
		Uint64Converter:  &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions: builtInFunctions.NewBuiltInFunctionContainer(),
	}
	return arguments
}

func TestNewVMContainerFactory_NilGasScheduleShouldErr(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		config.VirtualMachineConfig{},
		10000,
		nil,
		createMockVMAccountsArguments(),
	)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilGasSchedule, err)
}

func TestNewVMContainerFactory_OkValues(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		config.VirtualMachineConfig{},
		10000,
		coreConfig.MakeGasMapForTests(),
		createMockVMAccountsArguments(),
	)

	assert.NotNil(t, vmf)
	assert.Nil(t, err)
	assert.False(t, vmf.IsInterfaceNil())
}

func TestVmContainerFactory_Create(t *testing.T) {
	t.Parallel()

	vmf, err := NewVMContainerFactory(
		config.VirtualMachineConfig{},
		10000,
		coreConfig.MakeGasMapForTests(),
		createMockVMAccountsArguments(),
	)
	assert.NotNil(t, vmf)
	assert.Nil(t, err)

	container, err := vmf.Create()
	require.Nil(t, err)
	require.NotNil(t, container)
	defer func() {
		_ = container.Close()
	}()

	assert.Nil(t, err)
	assert.NotNil(t, container)

	vm, err := container.Get(factory.CoreVirtualMachine)
	assert.Nil(t, err)
	assert.NotNil(t, vm)

	acc := vmf.BlockChainHookImpl()
	assert.NotNil(t, acc)
}
