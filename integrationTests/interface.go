package integrationTests

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/epochStart"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// TestBootstrapper extends the Bootstrapper interface with some functions intended to be used only in tests
// as it simplifies the reproduction of edge cases
type TestBootstrapper interface {
	process.Bootstrapper
	RollBack(revertUsingForkNonce bool) error
	SetProbableHighestNonce(nonce uint64)
}

// TestEpochStartTrigger extends the epochStart trigger interface with some functions intended to by used only
// in tests as it simplifies the reproduction of test scenarios
type TestEpochStartTrigger interface {
	epochStart.TriggerHandler
	GetRoundsPerEpoch() uint64
	SetTrigger(triggerHandler epochStart.TriggerHandler)
	SetRoundsPerEpoch(roundsPerEpoch uint64)
	SetEpoch(epoch uint32)
}

// NodesCoordinatorFactory is used for creating a nodesCoordinator in the integration tests
type NodesCoordinatorFactory interface {
	CreateNodesCoordinator(arg ArgIndexHashedNodesCoordinatorFactory) sharding.NodesCoordinator
}

// NetworkShardingUpdater defines the updating methods used by the network sharding component
type NetworkShardingUpdater interface {
	GetPeerInfo(pid core.PeerID) core.P2PPeerInfo
	UpdatePeerIdPublicKey(pid core.PeerID, pk []byte)
	UpdatePublicKeyShardId(pk []byte, shardId uint32)
	UpdatePeerIdShardId(pid core.PeerID, shardId uint32)
	IsInterfaceNil() bool
}
