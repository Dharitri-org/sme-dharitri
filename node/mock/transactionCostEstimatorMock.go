package mock

import "github.com/Dharitri-org/sme-dharitri/data/transaction"

// TransactionCostEstimatorMock  --
type TransactionCostEstimatorMock struct {
	ComputeTransactionGasLimitCalled func(tx *transaction.Transaction) (uint64, error)
}

// ComputeTransactionGasLimit --
func (tcem *TransactionCostEstimatorMock) ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error) {
	if tcem.ComputeTransactionGasLimitCalled != nil {
		return tcem.ComputeTransactionGasLimitCalled(tx)
	}
	return 0, nil
}

// IsInterfaceNil --
func (tcem *TransactionCostEstimatorMock) IsInterfaceNil() bool {
	return tcem == nil
}
