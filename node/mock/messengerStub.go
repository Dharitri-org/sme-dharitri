package mock

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/p2p"
)

// MessengerStub -
type MessengerStub struct {
	IDCalled                         func() core.PeerID
	CloseCalled                      func() error
	CreateTopicCalled                func(name string, createChannelForTopic bool) error
	HasTopicCalled                   func(name string) bool
	HasTopicValidatorCalled          func(name string) bool
	BroadcastOnChannelCalled         func(channel string, topic string, buff []byte)
	BroadcastCalled                  func(topic string, buff []byte)
	RegisterMessageProcessorCalled   func(topic string, handler p2p.MessageProcessor) error
	BootstrapCalled                  func() error
	PeerAddressesCalled              func(pid core.PeerID) []string
	BroadcastOnChannelBlockingCalled func(channel string, topic string, buff []byte) error
	IsConnectedToTheNetworkCalled    func() bool
	PeersCalled                      func() []core.PeerID
}

// ID -
func (ms *MessengerStub) ID() core.PeerID {
	if ms.IDCalled != nil {
		return ms.IDCalled()
	}

	return ""
}

// RegisterMessageProcessor -
func (ms *MessengerStub) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	if ms.RegisterMessageProcessorCalled != nil {
		return ms.RegisterMessageProcessorCalled(topic, handler)
	}
	return nil
}

// Broadcast -
func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	ms.BroadcastCalled(topic, buff)
}

// Close -
func (ms *MessengerStub) Close() error {
	return ms.CloseCalled()
}

// CreateTopic -
func (ms *MessengerStub) CreateTopic(name string, createChannelForTopic bool) error {
	return ms.CreateTopicCalled(name, createChannelForTopic)
}

// HasTopic -
func (ms *MessengerStub) HasTopic(name string) bool {
	return ms.HasTopicCalled(name)
}

// HasTopicValidator -
func (ms *MessengerStub) HasTopicValidator(name string) bool {
	return ms.HasTopicValidatorCalled(name)
}

// BroadcastOnChannel -
func (ms *MessengerStub) BroadcastOnChannel(channel string, topic string, buff []byte) {
	ms.BroadcastOnChannelCalled(channel, topic, buff)
}

// Bootstrap -
func (ms *MessengerStub) Bootstrap() error {
	return ms.BootstrapCalled()
}

// PeerAddresses -
func (ms *MessengerStub) PeerAddresses(pid core.PeerID) []string {
	return ms.PeerAddressesCalled(pid)
}

// BroadcastOnChannelBlocking -
func (ms *MessengerStub) BroadcastOnChannelBlocking(channel string, topic string, buff []byte) error {
	return ms.BroadcastOnChannelBlockingCalled(channel, topic, buff)
}

// IsConnectedToTheNetwork -
func (ms *MessengerStub) IsConnectedToTheNetwork() bool {
	return ms.IsConnectedToTheNetworkCalled()
}

// Peers -
func (ms *MessengerStub) Peers() []core.PeerID {
	if ms.PeersCalled != nil {
		return ms.PeersCalled()
	}

	return make([]core.PeerID, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
