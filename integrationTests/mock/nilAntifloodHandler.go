package mock

import (
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/p2p"
	"github.com/Dharitri-org/sme-dharitri/process"
)

// NilAntifloodHandler is an empty implementation of P2PAntifloodHandler
// it does nothing
type NilAntifloodHandler struct {
}

// ResetForTopic won't do anything
func (nah *NilAntifloodHandler) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic won't do anything
func (nah *NilAntifloodHandler) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// CanProcessMessage will always return nil, allowing messages to go to interceptors
func (nah *NilAntifloodHandler) CanProcessMessage(_ p2p.MessageP2P, _ core.PeerID) error {
	return nil
}

// CanProcessMessagesOnTopic will always return nil, allowing messages to go to interceptors
func (nah *NilAntifloodHandler) CanProcessMessagesOnTopic(_ core.PeerID, _ string, _ uint32, _ uint64, _ []byte) error {
	return nil
}

// SetDebugger returns nil
func (nah *NilAntifloodHandler) SetDebugger(_ process.AntifloodDebugger) error {
	return nil
}

// ApplyConsensusSize does nothing
func (nah *NilAntifloodHandler) ApplyConsensusSize(_ int) {
}

// BlacklistPeer does nothing
func (nah *NilAntifloodHandler) BlacklistPeer(_ core.PeerID, _ string, _ time.Duration) {
}

// IsOriginatorEligibleForTopic returns nil
func (nah *NilAntifloodHandler) IsOriginatorEligibleForTopic(_ core.PeerID, _ string) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nah *NilAntifloodHandler) IsInterfaceNil() bool {
	return nah == nil
}
