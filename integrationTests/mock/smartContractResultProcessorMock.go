package mock

import (
	"github.com/Dharitri-org/sme-dharitri/data/smartContractResult"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

// SmartContractResultsProcessorMock -
type SmartContractResultsProcessorMock struct {
	ProcessSmartContractResultCalled func(scr *smartContractResult.SmartContractResult) error
}

// ProcessSmartContractResult -
func (scrp *SmartContractResultsProcessorMock) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	if scrp.ProcessSmartContractResultCalled == nil {
		return 0, nil
	}

	return 0, scrp.ProcessSmartContractResultCalled(scr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scrp *SmartContractResultsProcessorMock) IsInterfaceNil() bool {
	return scrp == nil
}
