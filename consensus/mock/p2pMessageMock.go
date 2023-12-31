package mock

import (
	"github.com/Dharitri-org/sme-dharitri/core"
)

// P2PMessageMock -
type P2PMessageMock struct {
	FromField      []byte
	DataField      []byte
	SeqNoField     []byte
	TopicsField    []string
	SignatureField []byte
	KeyField       []byte
	PeerField      core.PeerID
	PayloadField   []byte
	TimestampField int64
}

// From -
func (msg *P2PMessageMock) From() []byte {
	return msg.FromField
}

// Data -
func (msg *P2PMessageMock) Data() []byte {
	return msg.DataField
}

// SeqNo -
func (msg *P2PMessageMock) SeqNo() []byte {
	return msg.SeqNoField
}

// Topics -
func (msg *P2PMessageMock) Topics() []string {
	return msg.TopicsField
}

// Signature -
func (msg *P2PMessageMock) Signature() []byte {
	return msg.SignatureField
}

// Key -
func (msg *P2PMessageMock) Key() []byte {
	return msg.KeyField
}

// Peer -
func (msg *P2PMessageMock) Peer() core.PeerID {
	return msg.PeerField
}

// Timestamp -
func (msg *P2PMessageMock) Timestamp() int64 {
	return msg.TimestampField
}

// Payload -
func (msg *P2PMessageMock) Payload() []byte {
	return msg.PayloadField
}

// IsInterfaceNil returns true if there is no value under the interface
func (msg *P2PMessageMock) IsInterfaceNil() bool {
	return msg == nil
}
