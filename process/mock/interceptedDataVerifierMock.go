package mock

import "github.com/Dharitri-org/sme-dharitri/process"

type InterceptedDataVerifierMock struct {
}

// IsForCurrentShard -
func (i *InterceptedDataVerifierMock) IsForCurrentShard(_ process.InterceptedData) bool {
	return true
}

// IsInterfaceNil returns true if underlying object is
func (i *InterceptedDataVerifierMock) IsInterfaceNil() bool {
	return i == nil
}
