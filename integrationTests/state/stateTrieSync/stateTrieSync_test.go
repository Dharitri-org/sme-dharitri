package stateTrieSync

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/throttler"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/syncer"
	"github.com/Dharitri-org/sme-dharitri/data/trie"
	factory2 "github.com/Dharitri-org/sme-dharitri/data/trie/factory"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/requestHandlers"
	"github.com/Dharitri-org/sme-dharitri/integrationTests"
	"github.com/Dharitri-org/sme-dharitri/integrationTests/mock"
	"github.com/Dharitri-org/sme-dharitri/process/factory"
	"github.com/Dharitri-org/sme-dharitri/process/interceptors"
	"github.com/Dharitri-org/sme-dharitri/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_RequestInterceptTrieNodesWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0
	requesterNodeAddr := "0"
	resolverNodeAddr := "1"

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, requesterNodeAddr)
	_ = nRequester.Messenger.CreateTopic(core.ConsensusTopic+nRequester.ShardCoordinator.CommunicationIdentifier(nRequester.ShardCoordinator.SelfId()), true)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, resolverNodeAddr)
	_ = nResolver.Messenger.CreateTopic(core.ConsensusTopic+nResolver.ShardCoordinator.CommunicationIdentifier(nResolver.ShardCoordinator.SelfId()), true)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	resolverTrie := nResolver.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	//we have tested even with the 50000 value and found out that it worked in a reasonable amount of time ~21 seconds
	for i := 0; i < 10000; i++ {
		_ = resolverTrie.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	_ = resolverTrie.Commit()
	rootHash, _ := resolverTrie.Root()

	_, err = resolverTrie.GetAllLeaves()
	assert.Nil(t, err)

	requesterTrie := nRequester.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	nilRootHash, _ := requesterTrie.Root()
	whiteListHandler, _ := interceptors.NewWhiteListDataVerifier(
		&testscommon.CacherStub{
			PutCalled: func(_ []byte, _ interface{}, _ int) (evicted bool) {
				return false
			},
		},
	)
	requestHandler, _ := requestHandlers.NewResolverRequestHandler(
		nRequester.ResolverFinder,
		&mock.RequestedItemsHandlerStub{},
		whiteListHandler,
		10000,
		nRequester.ShardCoordinator.SelfId(),
		time.Second,
	)

	waitTime := 100 * time.Second
	trieSyncer, _ := trie.NewTrieSyncer(requestHandler, nRequester.DataPool.TrieNodes(), requesterTrie, shardID, factory.AccountTrieNodesTopic)
	ctx, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()

	err = trieSyncer.StartSyncing(rootHash, ctx)
	assert.Nil(t, err)

	newRootHash, _ := requesterTrie.Root()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)

	_, err = requesterTrie.GetAllLeaves()
	assert.Nil(t, err)
}

func TestMultipleDataTriesSync(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0
	requesterNodeAddr := "0"
	resolverNodeAddr := "1"

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, requesterNodeAddr)
	_ = nRequester.Messenger.CreateTopic(core.ConsensusTopic+nRequester.ShardCoordinator.CommunicationIdentifier(nRequester.ShardCoordinator.SelfId()), true)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, resolverNodeAddr)
	_ = nResolver.Messenger.CreateTopic(core.ConsensusTopic+nResolver.ShardCoordinator.CommunicationIdentifier(nResolver.ShardCoordinator.SelfId()), true)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	numAccounts := 1000
	numDataTrieLeaves := 50
	accState := nResolver.AccntState
	dataTrieRootHashes := make([][]byte, numAccounts)

	for i := 0; i < numAccounts; i++ {
		address := integrationTests.CreateAccount(accState, 1, big.NewInt(100))
		account, _ := accState.LoadAccount(address)
		userAcc, ok := account.(state.UserAccountHandler)
		assert.True(t, ok)

		rootHash := addValuesToDataTrie(t, accState, userAcc, numDataTrieLeaves)
		dataTrieRootHashes[i] = rootHash
	}

	rootHash, _ := accState.RootHash()
	_, err = accState.GetAllLeaves(rootHash)
	require.Nil(t, err)

	requesterTrie := nRequester.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	nilRootHash, _ := requesterTrie.Root()
	whiteListHandler, _ := interceptors.NewWhiteListDataVerifier(
		&testscommon.CacherStub{
			PutCalled: func(_ []byte, _ interface{}, _ int) (evicted bool) {
				return false
			},
		},
	)
	requestHandler, _ := requestHandlers.NewResolverRequestHandler(
		nRequester.ResolverFinder,
		&mock.RequestedItemsHandlerStub{},
		whiteListHandler,
		10000,
		nRequester.ShardCoordinator.SelfId(),
		time.Second,
	)

	thr, _ := throttler.NewNumGoRoutinesThrottler(50)
	syncerArgs := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:               integrationTests.TestHasher,
			Marshalizer:          integrationTests.TestMarshalizer,
			TrieStorageManager:   nRequester.TrieStorageManagers[factory2.UserAccountTrie],
			RequestHandler:       requestHandler,
			WaitTime:             time.Second * 300,
			Cacher:               nRequester.DataPool.TrieNodes(),
			MaxTrieLevelInMemory: 5,
		},
		ShardId:   shardID,
		Throttler: thr,
	}

	userAccSyncer, err := syncer.NewUserAccountsSyncer(syncerArgs)
	assert.Nil(t, err)

	err = userAccSyncer.SyncAccounts(rootHash)
	assert.Nil(t, err)

	_ = nRequester.AccntState.RecreateTrie(rootHash)

	newRootHash, _ := nRequester.AccntState.RootHash()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)

	leaves, err := nRequester.AccntState.GetAllLeaves(rootHash)
	assert.Nil(t, err)
	assert.Equal(t, numAccounts, len(leaves))
	checkAllDataTriesAreSynced(t, numDataTrieLeaves, nRequester.AccntState, dataTrieRootHashes)
}

func checkAllDataTriesAreSynced(t *testing.T, numDataTrieLeaves int, adb state.AccountsAdapter, dataTriesRootHashes [][]byte) {
	for i := range dataTriesRootHashes {
		dataTrieLeaves, err := adb.GetAllLeaves(dataTriesRootHashes[i])
		assert.Nil(t, err)
		assert.Equal(t, numDataTrieLeaves, len(dataTrieLeaves))
	}
}

func addValuesToDataTrie(t *testing.T, adb state.AccountsAdapter, acc state.UserAccountHandler, numVals int) []byte {
	for i := 0; i < numVals; i++ {
		randBytes := integrationTests.CreateRandomBytes(32)
		acc.DataTrieTracker().SaveKeyValue(randBytes, randBytes)
	}

	err := adb.SaveAccount(acc)
	assert.Nil(t, err)

	_, err = adb.Commit()
	assert.Nil(t, err)

	return acc.GetRootHash()
}
