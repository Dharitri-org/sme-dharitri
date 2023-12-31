package basicSync

import (
	"fmt"
	"testing"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/integrationTests"
	"github.com/stretchr/testify/assert"
)

func TestSyncWorksInShard_EmptyBlocksNoForks(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := uint32(1)
	shardId := uint32(0)
	numNodesPerShard := 6

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, numNodesPerShard+1)
	for i := 0; i < numNodesPerShard; i++ {
		nodes[i] = integrationTests.NewTestSyncNode(
			maxShards,
			shardId,
			shardId,
			advertiserAddr,
		)
	}

	metachainNode := integrationTests.NewTestSyncNode(
		maxShards,
		core.MetachainShardId,
		shardId,
		advertiserAddr,
	)
	idxProposerMeta := numNodesPerShard
	nodes[idxProposerMeta] = metachainNode

	idxProposerShard0 := 0
	idxProposers := []int{idxProposerShard0, idxProposerMeta}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
		_ = n.StartSync()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(nodes, round)
	nonce++

	numRoundsToTest := 5
	for i := 0; i < numRoundsToTest; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)

		time.Sleep(integrationTests.SyncDelay)

		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(nodes, round)
		nonce++
	}

	time.Sleep(integrationTests.SyncDelay)

	testAllNodesHaveTheSameBlockHeightInBlockchain(t, nodes)
}

func TestSyncWorksInShard_EmptyBlocksDoubleSign(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := uint32(1)
	shardId := uint32(0)
	numNodesPerShard := 6

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, numNodesPerShard)
	for i := 0; i < numNodesPerShard; i++ {
		nodes[i] = integrationTests.NewTestSyncNode(
			maxShards,
			shardId,
			shardId,
			advertiserAddr,
		)
	}

	idxProposerShard0 := 0
	idxProposers := []int{idxProposerShard0}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
		_ = n.StartSync()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(nodes, round)
	nonce++

	numRoundsToTest := 2
	for i := 0; i < numRoundsToTest; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)

		time.Sleep(integrationTests.SyncDelay)

		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.UpdateRound(nodes, round)
		nonce++
	}

	time.Sleep(integrationTests.SyncDelay)

	pubKeysVariant1 := []byte{3}
	pubKeysVariant2 := []byte{1}

	proposeBlockWithPubKeyBitmap(nodes[idxProposerShard0], round, nonce, pubKeysVariant1)
	proposeBlockWithPubKeyBitmap(nodes[1], round, nonce, pubKeysVariant2)

	time.Sleep(integrationTests.StepDelay)

	round = integrationTests.IncrementAndPrintRound(round)
	integrationTests.UpdateRound(nodes, round)

	stepDelayForkResolving := 4 * integrationTests.StepDelay
	time.Sleep(stepDelayForkResolving)

	testAllNodesHaveTheSameBlockHeightInBlockchain(t, nodes)
	testAllNodesHaveSameLastBlock(t, nodes)
}

func proposeBlockWithPubKeyBitmap(n *integrationTests.TestProcessorNode, round uint64, nonce uint64, pubKeys []byte) {
	body, header, _ := n.ProposeBlock(round, nonce)
	header.SetPubKeysBitmap(pubKeys)
	n.BroadcastBlock(body, header)
	n.CommitBlock(body, header)
}

func testAllNodesHaveTheSameBlockHeightInBlockchain(t *testing.T, nodes []*integrationTests.TestProcessorNode) {
	expectedNonce := nodes[0].BlockChain.GetCurrentBlockHeader().GetNonce()
	for i := 1; i < len(nodes); i++ {
		if check.IfNil(nodes[i].BlockChain.GetCurrentBlockHeader()) {
			assert.Fail(t, fmt.Sprintf("Node with idx %d does not have a current block", i))
		} else {
			assert.Equal(t, expectedNonce, nodes[i].BlockChain.GetCurrentBlockHeader().GetNonce())
		}
	}
}

func testAllNodesHaveSameLastBlock(t *testing.T, nodes []*integrationTests.TestProcessorNode) {
	mapBlocksByHash := make(map[string]data.HeaderHandler)

	for _, n := range nodes {
		hdr := n.BlockChain.GetCurrentBlockHeader()
		buff, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, hdr)

		mapBlocksByHash[string(buff)] = hdr
	}

	assert.Equal(t, 1, len(mapBlocksByHash))
}
