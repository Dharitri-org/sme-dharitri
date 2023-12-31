package mock

import (
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/process"
)

// BlockChainHookHandlerMock -
type BlockChainHookHandlerMock struct {
	SetCurrentHeaderCalled func(hdr data.HeaderHandler)
	NewAddressCalled       func(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error)
	IsPayableCalled        func(address []byte) (bool, error)
}

// IsPayable -
func (e *BlockChainHookHandlerMock) IsPayable(address []byte) (bool, error) {
	if e.IsPayableCalled != nil {
		return e.IsPayableCalled(address)
	}
	return true, nil
}

// GetBuiltInFunctions -
func (e *BlockChainHookHandlerMock) GetBuiltInFunctions() process.BuiltInFunctionContainer {
	return nil
}

// SetCurrentHeader -
func (e *BlockChainHookHandlerMock) SetCurrentHeader(hdr data.HeaderHandler) {
	if e.SetCurrentHeaderCalled != nil {
		e.SetCurrentHeaderCalled(hdr)
	}
}

// NewAddress -
func (e *BlockChainHookHandlerMock) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	if e.NewAddressCalled != nil {
		return e.NewAddressCalled(creatorAddress, creatorNonce, vmType)
	}

	return make([]byte, 0), nil
}

// IsInterfaceNil -
func (e *BlockChainHookHandlerMock) IsInterfaceNil() bool {
	return e == nil
}
