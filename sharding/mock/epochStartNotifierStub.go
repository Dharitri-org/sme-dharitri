package mock

import (
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/epochStart"
)

// EpochStartNotifierStub -
type EpochStartNotifierStub struct {
	RegisterHandlerCalled   func(handler epochStart.ActionHandler)
	UnregisterHandlerCalled func(handler epochStart.ActionHandler)
	NotifyAllPrepareCalled  func(hdr data.HeaderHandler, body data.BodyHandler)
	NotifyAllCalled         func(hdr data.HeaderHandler)
}

// RegisterHandler -
func (esnm *EpochStartNotifierStub) RegisterHandler(handler epochStart.ActionHandler) {
	if esnm.RegisterHandlerCalled != nil {
		esnm.RegisterHandlerCalled(handler)
	}
}

// UnregisterHandler -
func (esnm *EpochStartNotifierStub) UnregisterHandler(handler epochStart.ActionHandler) {
	if esnm.UnregisterHandlerCalled != nil {
		esnm.UnregisterHandlerCalled(handler)
	}
}

// NotifyAllPrepare -
func (esnm *EpochStartNotifierStub) NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler) {
	if esnm.NotifyAllPrepareCalled != nil {
		esnm.NotifyAllPrepareCalled(metaHdr, body)
	}
}

// NotifyAll -
func (esnm *EpochStartNotifierStub) NotifyAll(hdr data.HeaderHandler) {
	if esnm.NotifyAllCalled != nil {
		esnm.NotifyAllCalled(hdr)
	}
}

// IsInterfaceNil -
func (esnm *EpochStartNotifierStub) IsInterfaceNil() bool {
	return esnm == nil
}
