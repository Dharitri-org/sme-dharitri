package external

import (
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

// SCQueryService defines how data should be get from a SC account
type SCQueryService interface {
	ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error)
	ComputeScCallGasLimit(tx *transaction.Transaction) (uint64, error)
	IsInterfaceNil() bool
}

// StatusMetricsHandler is the interface that defines what a node details handler/provider should do
type StatusMetricsHandler interface {
	StatusMetricsMapWithoutP2P() map[string]interface{}
	StatusP2pMetricsMap() map[string]interface{}
	StatusMetricsWithoutP2PPrometheusString() string
	ConfigMetrics() map[string]interface{}
	NetworkMetrics() map[string]interface{}
	IsInterfaceNil() bool
}

// TransactionCostHandler defines the actions which should be handler by a transaction cost estimator
type TransactionCostHandler interface {
	ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error)
	IsInterfaceNil() bool
}
