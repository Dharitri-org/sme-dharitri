package mock

import (
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/data/smartContractResult"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

// TxProcessorMock -
type TxProcessorMock struct {
	ProcessTransactionCalled         func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error)
	SetBalancesToTrieCalled          func(accBalance map[string]*big.Int) (rootHash []byte, err error)
	ProcessSmartContractResultCalled func(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error)
}

// ProcessTransaction -
func (etm *TxProcessorMock) ProcessTransaction(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
	return etm.ProcessTransactionCalled(transaction)
}

// SetBalancesToTrie -
func (etm *TxProcessorMock) SetBalancesToTrie(accBalance map[string]*big.Int) (rootHash []byte, err error) {
	return etm.SetBalancesToTrieCalled(accBalance)
}

// ProcessSmartContractResult -
func (etm *TxProcessorMock) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	return etm.ProcessSmartContractResultCalled(scr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (etm *TxProcessorMock) IsInterfaceNil() bool {
	return etm == nil
}
