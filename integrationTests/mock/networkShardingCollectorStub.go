package mock

import (
	"github.com/Dharitri-org/sme-dharitri/core"
)

// NetworkShardingCollectorStub -
type NetworkShardingCollectorStub struct {
	UpdatePeerIdPublicKeyCalled  func(pid core.PeerID, pk []byte)
	UpdatePublicKeyShardIdCalled func(pk []byte, shardId uint32)
	UpdatePeerIdShardIdCalled    func(pid core.PeerID, shardId uint32)
}

// UpdatePeerIdPublicKey -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIdPublicKey(pid core.PeerID, pk []byte) {
	nscs.UpdatePeerIdPublicKeyCalled(pid, pk)
}

// UpdatePublicKeyShardId -
func (nscs *NetworkShardingCollectorStub) UpdatePublicKeyShardId(pk []byte, shardId uint32) {
	nscs.UpdatePublicKeyShardIdCalled(pk, shardId)
}

// UpdatePeerIdShardId -
func (nscs *NetworkShardingCollectorStub) UpdatePeerIdShardId(pid core.PeerID, shardId uint32) {
	nscs.UpdatePeerIdShardIdCalled(pid, shardId)
}

// IsInterfaceNil -
func (nscs *NetworkShardingCollectorStub) IsInterfaceNil() bool {
	return nscs == nil
}
