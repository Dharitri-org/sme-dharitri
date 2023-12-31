//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/Dharitri-org/protobuf/protobuf  --gogoslick_out=. dct.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/vm"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

const minLengthForTokenName = 10
const maxLengthForTokenName = 20
const configKeyPrefix = "dctConfig"
const burnable = "burnable"
const mintable = "mintable"
const canPause = "canPause"
const canFreeze = "canFreeze"
const canWipe = "canWipe"

const conversionBase = 10

type dct struct {
	eei             vm.SystemEI
	gasCost         vm.GasCost
	baseIssuingCost *big.Int
	ownerAddress    []byte
	dCTSCAddress    []byte
	marshalizer     marshal.Marshalizer
	hasher          hashing.Hasher
	disabled        bool
}

// ArgsNewDCTSmartContract defines the arguments needed for the dct contract
type ArgsNewDCTSmartContract struct {
	Eei          vm.SystemEI
	GasCost      vm.GasCost
	DCTSCConfig  config.DCTSystemSCConfig
	DCTSCAddress []byte
	Marshalizer  marshal.Marshalizer
	Hasher       hashing.Hasher
}

// NewDCTSmartContract creates the dct smart contract, which controls the issuing of tokens
func NewDCTSmartContract(args ArgsNewDCTSmartContract) (*dct, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, vm.ErrNilHasher
	}

	baseIssuingCost, ok := big.NewInt(0).SetString(args.DCTSCConfig.BaseIssuingCost, conversionBase)
	if !ok || baseIssuingCost.Cmp(big.NewInt(0)) < 0 {
		return nil, vm.ErrInvalidBaseIssuingCost
	}

	return &dct{
		eei:             args.Eei,
		gasCost:         args.GasCost,
		baseIssuingCost: baseIssuingCost,
		ownerAddress:    []byte(args.DCTSCConfig.OwnerAddress),
		dCTSCAddress:    args.DCTSCAddress,
		hasher:          args.Hasher,
		marshalizer:     args.Marshalizer,
		disabled:        args.DCTSCConfig.Disabled,
	}, nil
}

// Execute calls one of the functions from the dct smart contract and runs the code according to the input
func (e *dct) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	if args.Function == core.SCDeployInitFunctionName {
		return e.init(args)
	}

	if e.disabled {
		e.eei.AddReturnMessage("DCT SC disabled")
		return vmcommon.UserError
	}

	switch args.Function {
	case "issue":
		return e.issue(args)
	case "issueProtected":
		return e.issueProtected(args)
	case "burn":
		return e.burn(args)
	case "mint":
		return e.mint(args)
	case "freeze":
		return e.freeze(args)
	case "wipe":
		return e.wipe(args)
	case "pause":
		return e.pause(args)
	case "unPause":
		return e.unpause(args)
	case "claim":
		return e.claim(args)
	case "configChange":
		return e.configChange(args)
	case "dctControlChanges":
		return e.dctControlChanges(args)
	}

	e.eei.AddReturnMessage("invalid method to call")
	return vmcommon.FunctionNotFound
}

func (e *dct) init(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	scConfig := &DCTConfig{
		OwnerAddress:       e.ownerAddress,
		BaseIssuingCost:    e.baseIssuingCost,
		MinTokenNameLength: minLengthForTokenName,
		MaxTokenNameLength: maxLengthForTokenName,
	}
	marshaledData, err := e.marshalizer.Marshal(scConfig)
	log.LogIfError(err, "marshal error on dct init function")

	e.eei.SetStorage([]byte(configKeyPrefix), marshaledData)
	return vmcommon.Ok
}

func (e *dct) issueProtected(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, e.ownerAddress) {
		e.eei.AddReturnMessage("issueProtected can be called by whitelisted address only")
		return vmcommon.UserError
	}
	if len(args.Arguments) < 3 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	if len(args.Arguments[0]) < len(args.CallerAddr) {
		e.eei.AddReturnMessage("token name length not in parameters")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(e.baseIssuingCost) != 0 {
		e.eei.AddReturnMessage("callValue not equals with baseIssuingCost")
		return vmcommon.OutOfFunds
	}
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.DCTIssue)
	if err != nil {
		e.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}

	err = e.issueToken(args.Arguments[0], args.Arguments[1:])
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (e *dct) issue(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 2 {
		e.eei.AddReturnMessage("not enough arguments")
		return vmcommon.FunctionWrongSignature
	}
	if len(args.Arguments[0]) < minLengthForTokenName || len(args.Arguments[0]) > maxLengthForTokenName {
		e.eei.AddReturnMessage("token name length not in parameters")
		return vmcommon.FunctionWrongSignature
	}
	if args.CallValue.Cmp(e.baseIssuingCost) != 0 {
		e.eei.AddReturnMessage("callValue not equals with baseIssuingCost")
		return vmcommon.OutOfFunds
	}
	err := e.eei.UseGas(e.gasCost.MetaChainSystemSCsCost.DCTIssue)
	if err != nil {
		e.eei.AddReturnMessage("not enough gas")
		return vmcommon.OutOfGas
	}

	err = e.issueToken(args.CallerAddr, args.Arguments)
	if err != nil {
		e.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func isTokenNameHumanReadable(tokenName []byte) bool {
	for _, ch := range tokenName {
		isSmallCharacter := ch >= 'a' && ch <= 'z'
		isBigCharacter := ch >= 'A' && ch <= 'Z'
		isNumber := ch >= '0' && ch <= '9'
		isReadable := isSmallCharacter || isBigCharacter || isNumber
		if !isReadable {
			return false
		}
	}
	return true
}

func (e *dct) issueToken(owner []byte, arguments [][]byte) error {
	tokenName := arguments[0]
	initialSupply := big.NewInt(0).SetBytes(arguments[1])
	if initialSupply.Cmp(big.NewInt(0)) <= 0 {
		return vm.ErrNegativeOrZeroInitialSupply
	}

	data := e.eei.GetStorage(tokenName)
	if len(data) > 0 {
		return vm.ErrTokenAlreadyRegistered
	}

	if !isTokenNameHumanReadable(tokenName) {
		return vm.ErrTokenNameNotHumanReadable
	}

	newDCTToken := &DCTData{
		IssuerAddress: owner,
		TokenName:     tokenName,
		Mintable:      false,
		Burnable:      false,
		CanPause:      false,
		Paused:        false,
		CanFreeze:     false,
		CanWipe:       false,
		MintedValue:   initialSupply,
		BurntValue:    big.NewInt(0),
	}
	for i := 2; i < len(arguments); i++ {
		optionalArg := string(arguments[i])
		switch optionalArg {
		case burnable:
			newDCTToken.Burnable = true
		case mintable:
			newDCTToken.Mintable = true
		case canPause:
			newDCTToken.CanPause = true
		case canFreeze:
			newDCTToken.CanFreeze = true
		case canWipe:
			newDCTToken.CanWipe = true
		}
	}

	marshaledData, err := e.marshalizer.Marshal(newDCTToken)
	if err != nil {
		return err
	}

	e.eei.SetStorage(tokenName, marshaledData)

	dctTransferData := core.BuiltInFunctionDCTTransfer + "@" + hex.EncodeToString(tokenName) + "@" + hex.EncodeToString(initialSupply.Bytes())
	err = e.eei.Transfer(owner, e.dCTSCAddress, big.NewInt(0), []byte(dctTransferData), 0)
	if err != nil {
		return err
	}

	return nil
}

func (e *dct) burn(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *dct) mint(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *dct) freeze(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *dct) wipe(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *dct) pause(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *dct) unpause(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *dct) configChange(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *dct) claim(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

func (e *dct) dctControlChanges(_ *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	//TODO: implement me
	return vmcommon.Ok
}

// IsInterfaceNil returns true if underlying object is nil
func (e *dct) IsInterfaceNil() bool {
	return e == nil
}
