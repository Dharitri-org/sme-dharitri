package mock

import (
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/node/external"
	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

// ApiResolverStub -
type ApiResolverStub struct {
	ExecuteSCQueryHandler             func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	StatusMetricsHandler              func() external.StatusMetricsHandler
	ComputeTransactionGasLimitHandler func(tx *transaction.Transaction) (uint64, error)
}

// ExecuteSCQuery -
func (ars *ApiResolverStub) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return ars.ExecuteSCQueryHandler(query)
}

// StatusMetrics -
func (ars *ApiResolverStub) StatusMetrics() external.StatusMetricsHandler {
	return ars.StatusMetricsHandler()
}

// ComputeTransactionGasLimit -
func (ars *ApiResolverStub) ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error) {
	return ars.ComputeTransactionGasLimitHandler(tx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ars *ApiResolverStub) IsInterfaceNil() bool {
	return ars == nil
}
