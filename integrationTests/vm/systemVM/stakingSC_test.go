package systemVM

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/integrationTests"
	"github.com/Dharitri-org/sme-dharitri/integrationTests/multiShard/endOfEpoch"
	"github.com/Dharitri-org/sme-dharitri/vm/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStakingUnstakingAndUnboundingOnMultiShardEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)
	verifyInitialBalance(t, nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	///////////------- send stake tx and check sender's balance
	var txData string
	genesisBlock := nodes[0].GenesisBlocks[core.MetachainShardId]
	metaBlock := genesisBlock.(*block.MetaBlock)
	nodePrice := big.NewInt(0).Set(metaBlock.EpochStart.Economics.NodePrice)
	oneEncoded := hex.EncodeToString(big.NewInt(1).Bytes())
	for index, node := range nodes {
		pubKey := generateUniqueKey(index)
		txData = "stake" + "@" + oneEncoded + "@" + pubKey + "@" + hex.EncodeToString([]byte("msg"))
		integrationTests.CreateAndSendTransaction(node, nodePrice, factory.AuctionSCAddress, txData)
	}

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 10
	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	time.Sleep(time.Second)

	checkAccountsAfterStaking(t, nodes)

	/////////------ send unStake tx
	for index, node := range nodes {
		pubKey := generateUniqueKey(index)
		txData = "unStake" + "@" + pubKey
		integrationTests.CreateAndSendTransaction(node, big.NewInt(0), factory.AuctionSCAddress, txData)
	}

	time.Sleep(time.Second)

	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	/////////----- wait for unbond period
	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 10, nonce, round, idxProposers)

	////////----- send unBond
	for index, node := range nodes {
		pubKey := generateUniqueKey(index)
		txData = "unBond" + "@" + pubKey
		integrationTests.CreateAndSendTransaction(node, big.NewInt(0), factory.AuctionSCAddress, txData)
	}

	time.Sleep(time.Second)

	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	verifyUnbound(t, nodes)
}

func TestStakingUnstakingAndUnboundingOnMultiShardEnvironmentWithValidatorStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2
	shardConsensusGroupSize := 1
	metaConsensusGroupSize := 1

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		numMetachainNodes,
		numOfShards,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		integrationTests.GetConnectableAddress(advertiser),
	)

	nodes := make([]*integrationTests.TestProcessorNode, 0)
	idxProposers := make([]int, numOfShards+1)

	for _, nds := range nodesMap {
		nodes = append(nodes, nds...)
	}

	for _, nds := range nodesMap {
		idx, err := getNodeIndex(nodes, nds[0])
		require.Nil(t, err)

		idxProposers = append(idxProposers, idx)
	}

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, nds := range nodesMap {
		fmt.Println(integrationTests.MakeDisplayTable(nds))
	}

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	verifyInitialBalance(t, nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	///////////------- send stake tx and check sender's balance
	genesisBlock := nodes[0].GenesisBlocks[core.MetachainShardId]
	metaBlock := genesisBlock.(*block.MetaBlock)
	nodePrice := big.NewInt(0).Set(metaBlock.EpochStart.Economics.NodePrice)
	oneEncoded := hex.EncodeToString(big.NewInt(1).Bytes())
	var txData string
	for index, node := range nodes {
		pubKey := generateUniqueKey(index)
		txData = "stake" + "@" + oneEncoded + "@" + pubKey + "@" + hex.EncodeToString([]byte("msg"))
		integrationTests.CreateAndSendTransaction(node, nodePrice, factory.AuctionSCAddress, txData)
	}

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 10
	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	time.Sleep(time.Second)

	consumedBalance := big.NewInt(0).Add(big.NewInt(int64(len(txData))), big.NewInt(0).SetUint64(integrationTests.MinTxGasLimit))
	consumedBalance.Mul(consumedBalance, big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))

	checkAccountsAfterStaking(t, nodes)

	/////////------ send unStake tx
	for index, node := range nodes {
		pubKey := generateUniqueKey(index)
		txData = "unStake" + "@" + pubKey
		integrationTests.CreateAndSendTransaction(node, big.NewInt(0), factory.AuctionSCAddress, txData)
	}
	consumed := big.NewInt(0).Add(big.NewInt(0).SetUint64(integrationTests.MinTxGasLimit), big.NewInt(int64(len(txData))))
	consumed.Mul(consumed, big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))
	consumedBalance.Add(consumedBalance, consumed)

	time.Sleep(time.Second)

	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	/////////----- wait for unbound period
	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 10, nonce, round, idxProposers)

	////////----- send unBound
	for index, node := range nodes {
		pubKey := generateUniqueKey(index)
		txData = "unBond" + "@" + pubKey
		integrationTests.CreateAndSendTransaction(node, big.NewInt(0), factory.AuctionSCAddress, txData)
	}
	consumed = big.NewInt(0).Add(big.NewInt(0).SetUint64(integrationTests.MinTxGasLimit), big.NewInt(int64(len(txData))))
	consumed.Mul(consumed, big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))
	consumedBalance.Add(consumedBalance, consumed)

	time.Sleep(time.Second)

	integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)

	verifyUnbound(t, nodes)
}

func TestStakeWithRewardsAddressAndValidatorStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2
	shardConsensusGroupSize := 1
	metaConsensusGroupSize := 1

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorAndTxKeys(
		nodesPerShard,
		numMetachainNodes,
		numOfShards,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		integrationTests.GetConnectableAddress(advertiser),
	)

	nodes := make([]*integrationTests.TestProcessorNode, 0)
	idxProposers := make([]int, numOfShards+1)

	for _, nds := range nodesMap {
		nodes = append(nodes, nds...)
	}

	for _, nds := range nodesMap {
		idx, err := getNodeIndex(nodes, nds[0])
		assert.Nil(t, err)

		idxProposers = append(idxProposers, idx)
	}
	integrationTests.DisplayAndStartNodes(nodes)

	roundsPerEpoch := uint64(5)
	for _, node := range nodes {
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, node := range nodesMap {
		fmt.Println(integrationTests.MakeDisplayTable(node))
	}

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	rewardAccount := integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, 0)

	verifyInitialBalance(t, nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	var txData string
	for _, node := range nodes {
		txData = "changeRewardAddress" + "@" + hex.EncodeToString(rewardAccount.Address)
		integrationTests.CreateAndSendTransaction(node, big.NewInt(0), factory.AuctionSCAddress, txData)
	}

	nbBlocksToProduce := roundsPerEpoch * 3
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode

	for i := uint64(0); i < nbBlocksToProduce; i++ {
		for _, nodesSlice := range nodesMap {
			integrationTests.UpdateRound(nodesSlice, round)
			integrationTests.AddSelfNotarizedHeaderByMetachain(nodesSlice)
		}

		_, _, consensusNodes = integrationTests.AllShardsProposeBlock(round, nonce, nodesMap)
		indexesProposers := endOfEpoch.GetBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++

		time.Sleep(1 * time.Second)
	}

	rewardShardID := nodes[0].ShardCoordinator.ComputeId(rewardAccount.Address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != rewardShardID {
			continue
		}

		rwdAccount := getAccountFromAddrBytes(node.AccntState, rewardAccount.Address)
		assert.True(t, rwdAccount.GetBalance().Cmp(big.NewInt(0)) > 0)
	}
}

func getNodeIndex(nodeList []*integrationTests.TestProcessorNode, node *integrationTests.TestProcessorNode) (int, error) {
	for i := range nodeList {
		if node == nodeList[i] {
			return i, nil
		}
	}

	return 0, errors.New("no such node in list")
}

func verifyUnbound(t *testing.T, nodes []*integrationTests.TestProcessorNode) {
	expectedValue := big.NewInt(0).SetUint64(9999963900)
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				sndAcc := getAccountFromAddrBytes(helperNode.AccntState, node.OwnAccount.Address)
				require.True(t, sndAcc.GetBalance().Cmp(expectedValue) == 0)
				break
			}
		}
	}
}

func checkAccountsAfterStaking(t *testing.T, nodes []*integrationTests.TestProcessorNode) {
	expectedValue := big.NewInt(0).SetUint64(9999986910)
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {

				sndAcc := getAccountFromAddrBytes(helperNode.AccntState, node.OwnAccount.Address)
				require.True(t, sndAcc.GetBalance().Cmp(expectedValue) == 0)
				break
			}
		}
	}
}

func verifyInitialBalance(t *testing.T, nodes []*integrationTests.TestProcessorNode, initialVal *big.Int) {
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				sndAcc := getAccountFromAddrBytes(helperNode.AccntState, node.OwnAccount.Address)
				require.Equal(t, initialVal, sndAcc.GetBalance())
				break
			}
		}
	}
}

func getAccountFromAddrBytes(accState state.AccountsAdapter, address []byte) state.UserAccountHandler {
	sndrAcc, _ := accState.GetExistingAccount(address)

	sndAccSt, _ := sndrAcc.(state.UserAccountHandler)

	return sndAccSt
}

func generateUniqueKey(identifier int) string {
	neededLength := 192
	uniqueIdentifier := fmt.Sprintf("%d", identifier)
	return strings.Repeat("0", neededLength-len(uniqueIdentifier)) + uniqueIdentifier
}
