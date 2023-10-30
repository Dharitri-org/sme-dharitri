package mock

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/p2p"
	"github.com/Dharitri-org/sme-dharitri/process"
)

// InterceptorStub -
type InterceptorStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
}

// ProcessReceivedMessage -
func (is *InterceptorStub) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	return is.ProcessReceivedMessageCalled(message)
}

// SetInterceptedDebugHandler -
func (is *InterceptorStub) SetInterceptedDebugHandler(_ process.InterceptedDebugger) error {
	return nil
}

// RegisterHandler -
func (is *InterceptorStub) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (is *InterceptorStub) IsInterfaceNil() bool {
	return is == nil
}
