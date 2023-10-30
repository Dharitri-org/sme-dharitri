package node

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/p2p"
)

func (n *Node) CreateConsensusTopic(messageProcessor p2p.MessageProcessor) error {
	return n.createConsensusTopic(messageProcessor)
}

func (n *Node) ComputeTransactionStatus(tx data.TransactionHandler, isInPool bool) core.TransactionStatus {
	return n.computeTransactionStatus(tx, isInPool)
}
