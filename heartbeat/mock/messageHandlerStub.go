package mock

import (
	"github.com/Dharitri-org/sme-dharitri/heartbeat/data"
	"github.com/Dharitri-org/sme-dharitri/p2p"
)

// MessageHandlerStub -
type MessageHandlerStub struct {
	CreateHeartbeatFromP2PMessageCalled func(message p2p.MessageP2P) (*data.Heartbeat, error)
}

// IsInterfaceNil -
func (mhs *MessageHandlerStub) IsInterfaceNil() bool {
	return false
}

// CreateHeartbeatFromP2PMessage -
func (mhs *MessageHandlerStub) CreateHeartbeatFromP2PMessage(message p2p.MessageP2P) (*data.Heartbeat, error) {
	return mhs.CreateHeartbeatFromP2PMessageCalled(message)
}
