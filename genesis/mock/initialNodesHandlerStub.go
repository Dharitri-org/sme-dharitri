package mock

import "github.com/Dharitri-org/sme-dharitri/sharding"

// InitialNodesHandlerStub -
type InitialNodesHandlerStub struct {
	InitialNodesInfoCalled func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler)
	MinNumberOfNodesCalled func() uint32
}

// InitialNodesInfo -
func (inhs *InitialNodesHandlerStub) InitialNodesInfo() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
	if inhs.InitialNodesInfoCalled != nil {
		return inhs.InitialNodesInfoCalled()
	}

	return make(map[uint32][]sharding.GenesisNodeInfoHandler), make(map[uint32][]sharding.GenesisNodeInfoHandler)
}

// MinNumberOfNodes -
func (inhs *InitialNodesHandlerStub) MinNumberOfNodes() uint32 {
	if inhs.MinNumberOfNodesCalled != nil {
		return inhs.MinNumberOfNodesCalled()
	}

	return 0
}

// IsInterfaceNil -
func (inhs *InitialNodesHandlerStub) IsInterfaceNil() bool {
	return inhs == nil
}
