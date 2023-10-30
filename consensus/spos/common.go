package spos

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// GetConsensusTopicID will construct and return the topic ID based on shard coordinator
func GetConsensusTopicID(shardCoordinator sharding.Coordinator) string {
	return core.ConsensusTopic + shardCoordinator.CommunicationIdentifier(shardCoordinator.SelfId())
}
