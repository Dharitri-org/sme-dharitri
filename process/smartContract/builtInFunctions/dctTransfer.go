//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/Dharitri-org/protobuf/protobuf  --gogoslick_out=. dct.proto
package builtInFunctions

import (
	"encoding/hex"
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

const dctKeyIdentifier = "dct"

var _ process.BuiltinFunction = (*dctTransfer)(nil)

var zero = big.NewInt(0)

type dctTransfer struct {
	funcGasCost uint64
	marshalizer marshal.Marshalizer
	keyPrefix   []byte
}

// NewDCTTransferFunc returns the dct transfer built-in function component
func NewDCTTransferFunc(
	funcGasCost uint64,
	marshalizer marshal.Marshalizer,
) (*dctTransfer, error) {
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	e := &dctTransfer{
		funcGasCost: funcGasCost,
		marshalizer: marshalizer,
		keyPrefix:   []byte(core.DharitriProtectedKeyPrefix + dctKeyIdentifier),
	}

	return e, nil
}

// ProcessBuiltinFunction will transfer the underlying dct balance of the account
func (e *dctTransfer) ProcessBuiltinFunction(
	acntSnd, acntDst state.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	if vmInput == nil {
		return nil, process.ErrNilVmInput
	}
	if len(vmInput.Arguments) != 2 {
		return nil, process.ErrInvalidArguments
	}
	if vmInput.CallValue.Cmp(zero) != 0 {
		return nil, process.ErrBuiltInFunctionCalledWithValue
	}

	value := big.NewInt(0).SetBytes(vmInput.Arguments[1])
	if value.Cmp(zero) <= 0 {
		return nil, process.ErrNegativeValue
	}

	gasRemaining := uint64(0)
	dctTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	log.Trace("dctTransfer", "sender", vmInput.CallerAddr, "receiver", vmInput.RecipientAddr, "value", value, "token", dctTokenKey)

	if !check.IfNil(acntSnd) {
		// gas is paid only by sender
		if vmInput.GasProvided < e.funcGasCost {
			return nil, process.ErrNotEnoughGas
		}

		gasRemaining = vmInput.GasProvided - e.funcGasCost
		err := e.addToDCTBalance(acntSnd, dctTokenKey, big.NewInt(0).Neg(value))
		if err != nil {
			return nil, err
		}
	}

	vmOutput := &vmcommon.VMOutput{GasRemaining: gasRemaining}
	if !check.IfNil(acntDst) {
		err := e.addToDCTBalance(acntDst, dctTokenKey, value)
		if err != nil {
			return nil, err
		}

		return vmOutput, nil
	}

	if core.IsSmartContractAddress(vmInput.CallerAddr) {
		// cross-shard DCT transfer call through a smart contract - needs the storage update in order to create the smart contract result
		dctTransferTxData := core.BuiltInFunctionDCTTransfer + "@" + hex.EncodeToString(vmInput.Arguments[0]) + "@" + hex.EncodeToString(vmInput.Arguments[1])
		vmOutput.OutputAccounts = make(map[string]*vmcommon.OutputAccount)
		vmOutput.OutputAccounts[string(vmInput.RecipientAddr)] = &vmcommon.OutputAccount{
			Address:  vmInput.RecipientAddr,
			Data:     []byte(dctTransferTxData),
			CallType: vmcommon.AsynchronousCall,
		}
	}

	return vmOutput, nil
}

func (e *dctTransfer) addToDCTBalance(userAcnt state.UserAccountHandler, key []byte, value *big.Int) error {
	dctData, err := e.getDCTDataFromKey(userAcnt, key)
	if err != nil {
		return err
	}

	dctData.Value.Add(dctData.Value, value)
	if dctData.Value.Cmp(zero) < 0 {
		return process.ErrInsufficientFunds
	}

	marshaledData, err := e.marshalizer.Marshal(dctData)
	if err != nil {
		return err
	}

	log.Trace("dct after transfer", "addr", userAcnt.AddressBytes(), "value", dctData.Value, "tokenKey", key)
	userAcnt.DataTrieTracker().SaveKeyValue(key, marshaledData)

	return nil
}

func (e *dctTransfer) getDCTDataFromKey(userAcnt state.UserAccountHandler, key []byte) (*DCToken, error) {
	dctData := &DCToken{Value: big.NewInt(0)}
	marshaledData, err := userAcnt.DataTrieTracker().RetrieveValue(key)
	if err != nil || len(marshaledData) == 0 {
		return dctData, nil
	}

	err = e.marshalizer.Unmarshal(dctData, marshaledData)
	if err != nil {
		return nil, err
	}

	return dctData, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *dctTransfer) IsInterfaceNil() bool {
	return e == nil
}
