package mock

import (
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/p2p"
)

// P2PAntifloodHandlerStub -
type P2PAntifloodHandlerStub struct {
	CanProcessMessageCalled         func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	CanProcessMessagesOnTopicCalled func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error
	ApplyConsensusSizeCalled        func(size int)
	BlacklistPeerCalled             func(peer core.PeerID, reason string, duration time.Duration)
}

// ResetForTopic -
func (p2pahs *P2PAntifloodHandlerStub) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic -
func (p2pahs *P2PAntifloodHandlerStub) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// ApplyConsensusSize -
func (p2pahs *P2PAntifloodHandlerStub) ApplyConsensusSize(size int) {
	if p2pahs.ApplyConsensusSizeCalled != nil {
		p2pahs.ApplyConsensusSizeCalled(size)
	}
}

// CanProcessMessage -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if p2pahs.CanProcessMessageCalled == nil {
		return nil
	}

	return p2pahs.CanProcessMessageCalled(message, fromConnectedPeer)
}

// CanProcessMessagesOnTopic -
func (p2pahs *P2PAntifloodHandlerStub) CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
	if p2pahs.CanProcessMessagesOnTopicCalled == nil {
		return nil
	}

	return p2pahs.CanProcessMessagesOnTopicCalled(peer, topic, numMessages, totalSize, sequence)
}

// BlacklistPeer -
func (p2pahs *P2PAntifloodHandlerStub) BlacklistPeer(peer core.PeerID, reason string, duration time.Duration) {
	if p2pahs.BlacklistPeerCalled != nil {
		p2pahs.BlacklistPeerCalled(peer, reason, duration)
	}
}

// IsInterfaceNil -
func (p2pahs *P2PAntifloodHandlerStub) IsInterfaceNil() bool {
	return p2pahs == nil
}
