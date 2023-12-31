package transaction

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/node/external"
	"github.com/Dharitri-org/sme-dharitri/process"
)

type transactionCostEstimator struct {
	txTypeHandler      process.TxTypeHandler
	feeHandler         process.FeeHandler
	query              external.SCQueryService
	storePerByteCost   uint64
	compilePerByteCost uint64
}

// NewTransactionCostEstimator will create a new transaction cost estimator
func NewTransactionCostEstimator(
	txTypeHandler process.TxTypeHandler,
	feeHandler process.FeeHandler,
	query external.SCQueryService,
	gasSchedule map[string]map[string]uint64,
) (*transactionCostEstimator, error) {
	if check.IfNil(txTypeHandler) {
		return nil, process.ErrNilTxTypeHandler
	}
	if check.IfNil(feeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(query) {
		return nil, external.ErrNilSCQueryService
	}

	compileCost, storeCost := getOperationCost(gasSchedule)

	return &transactionCostEstimator{
		txTypeHandler:      txTypeHandler,
		feeHandler:         feeHandler,
		query:              query,
		storePerByteCost:   compileCost,
		compilePerByteCost: storeCost,
	}, nil
}

func getOperationCost(gasSchedule map[string]map[string]uint64) (uint64, uint64) {
	baseOpMap, ok := gasSchedule[core.BaseOperationCost]
	if !ok {
		return 0, 0
	}

	storeCost, ok := baseOpMap["StorePerByte"]
	if !ok {
		return 0, 0
	}

	compilerCost, ok := baseOpMap["CompilePerByte"]
	if !ok {
		return 0, 0
	}

	return storeCost, compilerCost
}

// ComputeTransactionGasLimit will calculate how many gas units a transaction will consume
func (tce *transactionCostEstimator) ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error) {
	txType := tce.txTypeHandler.ComputeTransactionType(tx)
	tx.GasPrice = 1

	switch txType {
	case process.MoveBalance:
		return tce.feeHandler.ComputeGasLimit(tx), nil
	case process.SCDeployment:
		return tce.computeScDeployGasLimit(tx)
	case process.SCInvoking:
		return tce.computeScCallGasLimit(tx)
	case process.BuiltInFunctionCall:
		return tce.computeScCallGasLimit(tx)
	default:
		return 0, process.ErrWrongTransaction
	}
}

func (tce *transactionCostEstimator) computeScDeployGasLimit(tx *transaction.Transaction) (uint64, error) {
	scDeployCost := uint64(len(tx.Data)) * (tce.storePerByteCost + tce.compilePerByteCost)
	baseCost := tce.feeHandler.ComputeGasLimit(tx)

	return baseCost + scDeployCost, nil
}

func (tce *transactionCostEstimator) computeScCallGasLimit(tx *transaction.Transaction) (uint64, error) {
	scCallGasLimit, err := tce.query.ComputeScCallGasLimit(tx)
	if err != nil {
		return 0, err
	}

	baseCost := tce.feeHandler.ComputeGasLimit(tx)
	return baseCost + scCallGasLimit, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tce *transactionCostEstimator) IsInterfaceNil() bool {
	return tce == nil
}
