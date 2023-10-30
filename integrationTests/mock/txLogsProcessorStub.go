package mock

import (
	"github.com/Dharitri-org/sme-dharitri/data"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

// TxLogsProcessorStub -
type TxLogsProcessorStub struct {
	GetLogCalled  func(txHash []byte) (data.LogHandler, error)
	SaveLogCalled func(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error
}

// GetLog -
func (txls *TxLogsProcessorStub) GetLog(txHash []byte) (data.LogHandler, error) {
	if txls.GetLogCalled != nil {
		return txls.GetLogCalled(txHash)
	}

	return nil, nil
}

// SaveLog -
func (txls *TxLogsProcessorStub) SaveLog(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error {
	if txls.SaveLogCalled != nil {
		return txls.SaveLogCalled(txHash, tx, vmLogs)
	}

	return nil
}

// IsInterfaceNil -
func (txls *TxLogsProcessorStub) IsInterfaceNil() bool {
	return txls == nil
}