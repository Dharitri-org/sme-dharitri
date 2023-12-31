package block

import (
	"time"

	"github.com/Dharitri-org/sme-dharitri/integrationTests"
)

// StepDelay -
var StepDelay = time.Second * 2

// GetBlockProposersIndexes -
func GetBlockProposersIndexes(
	consensusMap map[uint32][]*integrationTests.TestProcessorNode,
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
) map[uint32]int {

	indexProposer := make(map[uint32]int)

	for sh, testNodeList := range nodesMap {
		for k, testNode := range testNodeList {
			if consensusMap[sh][0] == testNode {
				indexProposer[sh] = k
			}
		}
	}

	return indexProposer
}
