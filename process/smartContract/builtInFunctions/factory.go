package builtInFunctions

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	logger "github.com/Dharitri-org/sme-logger"
	"github.com/mitchellh/mapstructure"
)

var log = logger.GetOrCreate("process/smartContract/builtInFunctions")

// ArgsCreateBuiltInFunctionContainer -
type ArgsCreateBuiltInFunctionContainer struct {
	GasMap               map[string]map[string]uint64
	MapDNSAddresses      map[string]struct{}
	EnableUserNameChange bool
	Marshalizer          marshal.Marshalizer
}

// CreateBuiltInFunctionContainer will create the list of built-in functions
func CreateBuiltInFunctionContainer(args ArgsCreateBuiltInFunctionContainer) (process.BuiltInFunctionContainer, error) {
	gasConfig, err := createGasConfig(args.GasMap)
	if err != nil {
		return nil, err
	}

	container := NewBuiltInFunctionContainer()
	var newFunc process.BuiltinFunction
	newFunc = NewClaimDeveloperRewardsFunc(gasConfig.BuiltInCost.ClaimDeveloperRewards)
	err = container.Add(core.BuiltInFunctionClaimDeveloperRewards, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc = NewChangeOwnerAddressFunc(gasConfig.BuiltInCost.ChangeOwnerAddress)
	err = container.Add(core.BuiltInFunctionChangeOwnerAddress, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewSaveUserNameFunc(gasConfig.BuiltInCost.SaveUserName, args.MapDNSAddresses, args.EnableUserNameChange)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionSetUserName, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewSaveKeyValueStorageFunc(gasConfig.BaseOperationCost, gasConfig.BuiltInCost.SaveKeyValue)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionSaveKeyValue, newFunc)
	if err != nil {
		return nil, err
	}

	newFunc, err = NewDCTTransferFunc(gasConfig.BuiltInCost.DCTTransfer, args.Marshalizer)
	if err != nil {
		return nil, err
	}
	err = container.Add(core.BuiltInFunctionDCTTransfer, newFunc)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func createGasConfig(gasMap map[string]map[string]uint64) (*GasCost, error) {
	baseOps := &BaseOperationCost{}
	err := mapstructure.Decode(gasMap[core.BaseOperationCost], baseOps)
	if err != nil {
		return nil, err
	}

	err = check.ForZeroUintFields(*baseOps)
	if err != nil {
		return nil, err
	}

	builtInOps := &BuiltInCost{}
	err = mapstructure.Decode(gasMap[core.BuiltInCost], builtInOps)
	if err != nil {
		return nil, err
	}

	err = check.ForZeroUintFields(*builtInOps)
	if err != nil {
		return nil, err
	}

	gasCost := GasCost{
		BaseOperationCost: *baseOps,
		BuiltInCost:       *builtInOps,
	}

	return &gasCost, nil
}
