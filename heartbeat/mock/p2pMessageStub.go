package mock

import (
	"github.com/Dharitri-org/sme-dharitri/core"
)

// P2PMessageStub -
type P2PMessageStub struct {
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
func (msg *P2PMessageStub) From() []byte {
	return msg.FromField
}

// Data -
func (msg *P2PMessageStub) Data() []byte {
	return msg.DataField
}

// SeqNo -
func (msg *P2PMessageStub) SeqNo() []byte {
	return msg.SeqNoField
}

// Topics -
func (msg *P2PMessageStub) Topics() []string {
	return msg.TopicsField
}

// Signature -
func (msg *P2PMessageStub) Signature() []byte {
	return msg.SignatureField
}

// Key -
func (msg *P2PMessageStub) Key() []byte {
	return msg.KeyField
}

// Peer -
func (msg *P2PMessageStub) Peer() core.PeerID {
	return msg.PeerField
}

// Timestamp -
func (msg *P2PMessageStub) Timestamp() int64 {
	return msg.TimestampField
}

// Payload -
func (msg *P2PMessageStub) Payload() []byte {
	return msg.PayloadField
}

// IsInterfaceNil returns true if there is no value under the interface
func (msg *P2PMessageStub) IsInterfaceNil() bool {
	return msg == nil
}
