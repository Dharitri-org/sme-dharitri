package mock

import (
	"github.com/Dharitri-org/sme-dharitri/core"
)

// PeerListCreatorStub -
type PeerListCreatorStub struct {
	PeerListCalled           func() []core.PeerID
	IntraShardPeerListCalled func() []core.PeerID
}

// PeerList -
func (p *PeerListCreatorStub) PeerList() []core.PeerID {
	return p.PeerListCalled()
}

// IntraShardPeerList -
func (p *PeerListCreatorStub) IntraShardPeerList() []core.PeerID {
	return p.IntraShardPeerListCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *PeerListCreatorStub) IsInterfaceNil() bool {
	return p == nil
}
