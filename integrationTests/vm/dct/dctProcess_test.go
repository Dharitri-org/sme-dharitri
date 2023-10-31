package dct

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/integrationTests"
	"github.com/Dharitri-org/sme-dharitri/process/smartContract/builtInFunctions"
	"github.com/Dharitri-org/sme-dharitri/vm/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDCTIssueAndTransactionsOnMultiShardEnvironment(t *testing.T) {
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

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	///////////------- send token issue
	tokenName := "robertWhyNot"
	issuePrice := big.NewInt(1000)
	initalSupply := big.NewInt(10000000000)
	tokenIssuer := nodes[0]
	txData := "issue" + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(initalSupply.Bytes())
	integrationTests.CreateAndSendTransaction(tokenIssuer, issuePrice, factory.DCTSCAddress, txData)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasDCTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenName, initalSupply)

	/////////------ send tx to other nodes
	valueToSend := big.NewInt(100)
	for _, node := range nodes[1:] {
		txData = core.BuiltInFunctionDCTTransfer + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(valueToSend.Bytes())
		integrationTests.CreateAndSendTransaction(tokenIssuer, big.NewInt(0), node.OwnAccount.Address, txData)
	}

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	finalSupply := big.NewInt(0).Set(initalSupply)
	for _, node := range nodes[1:] {
		checkAddressHasDCTTokens(t, node.OwnAccount.Address, nodes, tokenName, valueToSend)
		finalSupply.Sub(finalSupply, valueToSend)
	}

	checkAddressHasDCTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenName, finalSupply)
}

func checkAddressHasDCTTokens(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	value *big.Int,
) {
	userAcc := getUserAccountWithAddress(t, address, nodes)
	require.False(t, check.IfNil(userAcc))

	tokenKey := []byte(core.DharitriProtectedKeyPrefix + "dct" + tokenName)
	dctData, err := getDCTDataFromKey(userAcc, tokenKey)
	assert.Nil(t, err)

	assert.Equal(t, dctData.Value.Cmp(value), 0)
}

func getUserAccountWithAddress(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
) state.UserAccountHandler {
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				acc, err := helperNode.AccntState.LoadAccount(address)
				require.Nil(t, err)
				return acc.(state.UserAccountHandler)
			}
		}
	}

	return nil
}

func getDCTDataFromKey(userAcnt state.UserAccountHandler, key []byte) (*builtInFunctions.DCToken, error) {
	dctData := &builtInFunctions.DCToken{Value: big.NewInt(0)}
	marshaledData, err := userAcnt.DataTrieTracker().RetrieveValue(key)
	if err != nil {
		return dctData, nil
	}

	err = integrationTests.TestMarshalizer.Unmarshal(dctData, marshaledData)
	if err != nil {
		return nil, err
	}

	return dctData, nil
}
