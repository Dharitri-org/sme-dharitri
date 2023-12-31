package factory

import (
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/vm"
	"github.com/Dharitri-org/sme-dharitri/vm/systemSmartContracts"
	"github.com/mitchellh/mapstructure"
)

type systemSCFactory struct {
	systemEI            vm.ContextHandler
	economics           vm.EconomicsHandler
	nodesConfigProvider vm.NodesConfigProvider
	sigVerifier         vm.MessageSignVerifier
	gasCost             vm.GasCost
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	systemSCConfig      *config.SystemSmartContractsConfig
}

// ArgsNewSystemSCFactory defines the arguments struct needed to create the system SCs
type ArgsNewSystemSCFactory struct {
	SystemEI            vm.ContextHandler
	Economics           vm.EconomicsHandler
	NodesConfigProvider vm.NodesConfigProvider
	SigVerifier         vm.MessageSignVerifier
	GasMap              map[string]map[string]uint64
	Marshalizer         marshal.Marshalizer
	Hasher              hashing.Hasher
	SystemSCConfig      *config.SystemSmartContractsConfig
}

// NewSystemSCFactory creates a factory which will instantiate the system smart contracts
func NewSystemSCFactory(args ArgsNewSystemSCFactory) (*systemSCFactory, error) {
	if check.IfNil(args.SystemEI) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if check.IfNil(args.SigVerifier) {
		return nil, vm.ErrNilMessageSignVerifier
	}
	if check.IfNil(args.NodesConfigProvider) {
		return nil, vm.ErrNilNodesConfigProvider
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, vm.ErrNilHasher
	}
	if check.IfNil(args.Economics) {
		return nil, vm.ErrNilEconomicsData
	}
	if args.SystemSCConfig == nil {
		return nil, vm.ErrNilSystemSCConfig
	}

	scf := &systemSCFactory{
		systemEI:            args.SystemEI,
		sigVerifier:         args.SigVerifier,
		nodesConfigProvider: args.NodesConfigProvider,
		marshalizer:         args.Marshalizer,
		hasher:              args.Hasher,
		systemSCConfig:      args.SystemSCConfig,
		economics:           args.Economics,
	}

	err := scf.createGasConfig(args.GasMap)
	if err != nil {
		return nil, err
	}

	return scf, nil
}

func (scf *systemSCFactory) createGasConfig(gasMap map[string]map[string]uint64) error {
	baseOps := &vm.BaseOperationCost{}
	err := mapstructure.Decode(gasMap[core.BaseOperationCost], baseOps)
	if err != nil {
		return err
	}

	err = check.ForZeroUintFields(*baseOps)
	if err != nil {
		return err
	}

	metaChainSCsOps := &vm.MetaChainSystemSCsCost{}
	err = mapstructure.Decode(gasMap[core.MetaChainSystemSCsCost], metaChainSCsOps)
	if err != nil {
		return err
	}

	err = check.ForZeroUintFields(*metaChainSCsOps)
	if err != nil {
		return err
	}

	builtInFunctionsCost := &vm.BuiltInCost{}
	err = mapstructure.Decode(gasMap[core.BuiltInCost], builtInFunctionsCost)
	if err != nil {
		return err
	}

	scf.gasCost = vm.GasCost{
		BaseOperationCost:      *baseOps,
		MetaChainSystemSCsCost: *metaChainSCsOps,
		BuiltInCost:            *builtInFunctionsCost,
	}

	return nil
}

func (scf *systemSCFactory) createStakingContract() (vm.SystemSmartContract, error) {
	argsStaking := systemSmartContracts.ArgsNewStakingSmartContract{
		MinNumNodes:       uint64(scf.nodesConfigProvider.MinNumberOfNodes()),
		StakingSCConfig:   scf.systemSCConfig.StakingSystemSCConfig,
		Eei:               scf.systemEI,
		StakingAccessAddr: AuctionSCAddress,
		JailAccessAddr:    JailingAddress,
		GasCost:           scf.gasCost,
		Marshalizer:       scf.marshalizer,
	}
	staking, err := systemSmartContracts.NewStakingSmartContract(argsStaking)
	return staking, err
}

func (scf *systemSCFactory) createAuctionContract() (vm.SystemSmartContract, error) {
	args := systemSmartContracts.ArgsStakingAuctionSmartContract{
		Eei:                scf.systemEI,
		SigVerifier:        scf.sigVerifier,
		StakingSCConfig:    scf.systemSCConfig.StakingSystemSCConfig,
		StakingSCAddress:   StakingSCAddress,
		AuctionSCAddress:   AuctionSCAddress,
		GasCost:            scf.gasCost,
		Marshalizer:        scf.marshalizer,
		NumOfNodesToSelect: scf.systemSCConfig.StakingSystemSCConfig.NodesToSelectInAuction,
		GenesisTotalSupply: scf.economics.GenesisTotalSupply(),
	}
	auction, err := systemSmartContracts.NewStakingAuctionSmartContract(args)
	return auction, err
}

func (scf *systemSCFactory) createDCTContract() (vm.SystemSmartContract, error) {
	argsDCT := systemSmartContracts.ArgsNewDCTSmartContract{
		Eei:          scf.systemEI,
		GasCost:      scf.gasCost,
		DCTSCAddress: DCTSCAddress,
		Marshalizer:  scf.marshalizer,
		Hasher:       scf.hasher,
		DCTSCConfig:  scf.systemSCConfig.DCTSystemSCConfig,
	}
	dct, err := systemSmartContracts.NewDCTSmartContract(argsDCT)
	return dct, err
}

func (scf *systemSCFactory) createGovernanceContract() (vm.SystemSmartContract, error) {
	argsGovernance := systemSmartContracts.ArgsNewGovernanceContract{
		Eei:                 scf.systemEI,
		GasCost:             scf.gasCost,
		GovernanceConfig:    scf.systemSCConfig.GovernanceSystemSCConfig,
		DCTSCAddress:        DCTSCAddress,
		Marshalizer:         scf.marshalizer,
		Hasher:              scf.hasher,
		GovernanceSCAddress: GovernanceSCAddress,
		StakingSCAddress:    StakingSCAddress,
		AuctionSCAddress:    AuctionSCAddress,
	}
	governance, err := systemSmartContracts.NewGovernanceContract(argsGovernance)
	return governance, err
}

// Create instantiates all the system smart contracts and returns a container
func (scf *systemSCFactory) Create() (vm.SystemSCContainer, error) {
	scContainer := NewSystemSCContainer()

	staking, err := scf.createStakingContract()
	if err != nil {
		return nil, err
	}

	err = scContainer.Add(StakingSCAddress, staking)
	if err != nil {
		return nil, err
	}

	auction, err := scf.createAuctionContract()
	if err != nil {
		return nil, err
	}

	err = scContainer.Add(AuctionSCAddress, auction)
	if err != nil {
		return nil, err
	}

	dct, err := scf.createDCTContract()
	if err != nil {
		return nil, err
	}

	err = scContainer.Add(DCTSCAddress, dct)
	if err != nil {
		return nil, err
	}

	governance, err := scf.createGovernanceContract()
	if err != nil {
		return nil, err
	}

	err = scContainer.Add(GovernanceSCAddress, governance)
	if err != nil {
		return nil, err
	}

	err = scf.systemEI.SetSystemSCContainer(scContainer)
	if err != nil {
		return nil, err
	}

	return scContainer, nil
}

// IsInterfaceNil checks whether the underlying object is nil
func (scf *systemSCFactory) IsInterfaceNil() bool {
	return scf == nil
}
