package mock

import (
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

// QueryServiceStub -
type QueryServiceStub struct {
	ComputeScCallGasLimitCalled func(tx *transaction.Transaction) (uint64, error)
	ExecuteQueryCalled          func(query *process.SCQuery) (*vmcommon.VMOutput, error)
}

// ComputeScCallGasLimit -
func (qss *QueryServiceStub) ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error) {
	if qss.ComputeScCallGasLimitCalled != nil {
		return qss.ComputeScCallGasLimitCalled(tx)
	}

	return 0, nil
}

// ExecuteQuery -
func (qss *QueryServiceStub) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	if qss.ExecuteQueryCalled != nil {
		return qss.ExecuteQueryCalled(query)
	}

	return &vmcommon.VMOutput{}, nil
}

// IsInterfaceNil -
func (qss *QueryServiceStub) IsInterfaceNil() bool {
	return qss == nil
}
