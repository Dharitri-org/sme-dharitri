package node

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	syncGo "sync"
	"sync/atomic"
	"time"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/consensus"
	"github.com/Dharitri-org/sme-dharitri/consensus/chronology"
	"github.com/Dharitri-org/sme-dharitri/consensus/spos"
	"github.com/Dharitri-org/sme-dharitri/consensus/spos/sposFactory"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/core/fullHistory"
	"github.com/Dharitri-org/sme-dharitri/core/indexer"
	"github.com/Dharitri-org/sme-dharitri/core/partitioning"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/endProcess"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/data/typeConverters"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/provider"
	"github.com/Dharitri-org/sme-dharitri/debug"
	"github.com/Dharitri-org/sme-dharitri/epochStart"
	"github.com/Dharitri-org/sme-dharitri/facade"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/heartbeat/componentHandler"
	heartbeatData "github.com/Dharitri-org/sme-dharitri/heartbeat/data"
	heartbeatProcess "github.com/Dharitri-org/sme-dharitri/heartbeat/process"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/ntp"
	"github.com/Dharitri-org/sme-dharitri/p2p"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/dataValidators"
	"github.com/Dharitri-org/sme-dharitri/process/factory"
	"github.com/Dharitri-org/sme-dharitri/process/smartContract"
	"github.com/Dharitri-org/sme-dharitri/process/sync"
	"github.com/Dharitri-org/sme-dharitri/process/sync/storageBootstrap"
	procTx "github.com/Dharitri-org/sme-dharitri/process/transaction"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/statusHandler"
	"github.com/Dharitri-org/sme-dharitri/update"
	logger "github.com/Dharitri-org/sme-logger"
)

// SendTransactionsPipe is the pipe used for sending new transactions
const SendTransactionsPipe = "send transactions pipe"

var log = logger.GetOrCreate("node")
var numSecondsBetweenPrints = 20

var _ facade.NodeHandler = (*Node)(nil)

// Option represents a functional configuration parameter that can operate
//
//	over the None struct.
type Option func(*Node) error

// Node is a structure that passes the configuration parameters and initializes
//
//	required services as requested
type Node struct {
	internalMarshalizer           marshal.Marshalizer
	vmMarshalizer                 marshal.Marshalizer
	txSignMarshalizer             marshal.Marshalizer
	ctx                           context.Context
	hasher                        hashing.Hasher
	feeHandler                    process.FeeHandler
	initialNodesPubkeys           map[uint32][]string
	roundDuration                 uint64
	consensusGroupSize            int
	messenger                     P2PMessenger
	syncTimer                     ntp.SyncTimer
	rounder                       consensus.Rounder
	blockProcessor                process.BlockProcessor
	genesisTime                   time.Time
	epochStartTrigger             epochStart.TriggerHandler
	epochStartRegistrationHandler epochStart.RegistrationHandler
	accounts                      state.AccountsAdapter
	addressPubkeyConverter        core.PubkeyConverter
	validatorPubkeyConverter      core.PubkeyConverter
	uint64ByteSliceConverter      typeConverters.Uint64ByteSliceConverter
	interceptorsContainer         process.InterceptorsContainer
	resolversFinder               dataRetriever.ResolversFinder
	peerDenialEvaluator           p2p.PeerDenialEvaluator
	appStatusHandler              core.AppStatusHandler
	validatorStatistics           process.ValidatorStatisticsProcessor
	hardforkTrigger               HardforkTrigger
	validatorsProvider            process.ValidatorsProvider
	whiteListRequest              process.WhiteListHandler
	whiteListerVerifiedTxs        process.WhiteListHandler

	pubKey            crypto.PublicKey
	privKey           crypto.PrivateKey
	keyGen            crypto.KeyGenerator
	keyGenForAccounts crypto.KeyGenerator
	singleSigner      crypto.SingleSigner
	txSingleSigner    crypto.SingleSigner
	multiSigner       crypto.MultiSigner
	peerSigHandler    crypto.PeerSignatureHandler
	forkDetector      process.ForkDetector

	blkc               data.ChainHandler
	dataPool           dataRetriever.PoolsHolder
	store              dataRetriever.StorageService
	shardCoordinator   sharding.Coordinator
	nodesCoordinator   sharding.NodesCoordinator
	miniblocksProvider process.MiniBlockProvider

	networkShardingCollector NetworkShardingCollector

	consensusTopic string
	consensusType  string

	currentSendingGoRoutines int32
	bootstrapRoundIndex      uint64

	indexer                 indexer.Indexer
	blocksBlackListHandler  process.TimeCacher
	bootStorer              process.BootStorer
	requestedItemsHandler   dataRetriever.RequestedItemsHandler
	headerSigVerifier       spos.RandSeedVerifier
	headerIntegrityVerifier spos.HeaderIntegrityVerifier

	chainID               []byte
	minTransactionVersion uint32

	sizeCheckDelta        uint32
	txSentCounter         uint32
	inputAntifloodHandler P2PAntifloodHandler
	txAcumulator          Accumulator

	blockTracker             process.BlockTracker
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler

	requestHandler process.RequestHandler

	signatureSize int
	publicKeySize int

	chanStopNodeProcess chan endProcess.ArgEndProcess

	mutQueryHandlers syncGo.RWMutex
	queryHandlers    map[string]debug.QueryHandler

	heartbeatHandler   *componentHandler.HeartbeatHandler
	peerHonestyHandler consensus.PeerHonestyHandler

	watchdog          core.WatchdogTimer
	historyRepository fullHistory.HistoryRepository
}

// ApplyOptions can set up different configurable options of a Node instance
func (n *Node) ApplyOptions(opts ...Option) error {
	for _, opt := range opts {
		err := opt(n)
		if err != nil {
			return errors.New("error applying option: " + err.Error())
		}
	}
	return nil
}

// NewNode creates a new Node instance
func NewNode(opts ...Option) (*Node, error) {
	node := &Node{
		ctx:                      context.Background(),
		currentSendingGoRoutines: 0,
		appStatusHandler:         statusHandler.NewNilStatusHandler(),
		queryHandlers:            make(map[string]debug.QueryHandler),
	}
	for _, opt := range opts {
		err := opt(node)
		if err != nil {
			return nil, errors.New("error applying option: " + err.Error())
		}
	}

	return node, nil
}

// GetAppStatusHandler will return the current status handler
func (n *Node) GetAppStatusHandler() core.AppStatusHandler {
	return n.appStatusHandler
}

// CreateShardedStores instantiate sharded cachers for Transactions and Headers
func (n *Node) CreateShardedStores() error {
	if n.shardCoordinator == nil {
		return ErrNilShardCoordinator
	}

	if n.dataPool == nil {
		return ErrNilDataPool
	}

	transactionsDataStore := n.dataPool.Transactions()
	headersDataStore := n.dataPool.Headers()

	if transactionsDataStore == nil {
		return errors.New("nil transaction sharded data store")
	}

	if headersDataStore == nil {
		return errors.New("nil header sharded data store")
	}

	return nil
}

// StartConsensus will start the consensus service for the current node
func (n *Node) StartConsensus() error {
	isGenesisBlockNotInitialized := len(n.blkc.GetGenesisHeaderHash()) == 0 ||
		check.IfNil(n.blkc.GetGenesisHeader())
	if isGenesisBlockNotInitialized {
		return ErrGenesisBlockNotInitialized
	}

	chronologyHandler, err := n.createChronologyHandler(
		n.rounder,
		n.appStatusHandler,
		n.watchdog,
	)
	if err != nil {
		return err
	}

	bootstrapper, err := n.createBootstrapper(n.rounder)
	if err != nil {
		return err
	}

	err = bootstrapper.SetStatusHandler(n.GetAppStatusHandler())
	if err != nil {
		log.Debug("cannot set app status handler for shard bootstrapper")
	}

	bootstrapper.StartSyncingBlocks()

	epoch := n.blkc.GetGenesisHeader().GetEpoch()
	crtBlockHeader := n.blkc.GetCurrentBlockHeader()
	if !check.IfNil(crtBlockHeader) {
		epoch = crtBlockHeader.GetEpoch()
	}
	log.Info("starting consensus", "epoch", epoch)

	consensusState, err := n.createConsensusState(epoch)
	if err != nil {
		return err
	}

	consensusService, err := sposFactory.GetConsensusCoreFactory(n.consensusType)
	if err != nil {
		return err
	}

	broadcastMessenger, err := sposFactory.GetBroadcastMessenger(
		n.internalMarshalizer,
		n.hasher,
		n.messenger,
		n.shardCoordinator,
		n.privKey,
		n.peerSigHandler,
		n.dataPool.Headers(),
		n.interceptorsContainer,
	)

	if err != nil {
		return err
	}

	netInputMarshalizer := n.internalMarshalizer
	if n.sizeCheckDelta > 0 {
		netInputMarshalizer = marshal.NewSizeCheckUnmarshalizer(n.internalMarshalizer, n.sizeCheckDelta)
	}

	workerArgs := &spos.WorkerArgs{
		ConsensusService:         consensusService,
		BlockChain:               n.blkc,
		BlockProcessor:           n.blockProcessor,
		Bootstrapper:             bootstrapper,
		BroadcastMessenger:       broadcastMessenger,
		ConsensusState:           consensusState,
		ForkDetector:             n.forkDetector,
		Marshalizer:              netInputMarshalizer,
		Hasher:                   n.hasher,
		Rounder:                  n.rounder,
		ShardCoordinator:         n.shardCoordinator,
		PeerSignatureHandler:     n.peerSigHandler,
		SyncTimer:                n.syncTimer,
		HeaderSigVerifier:        n.headerSigVerifier,
		HeaderIntegrityVerifier:  n.headerIntegrityVerifier,
		ChainID:                  n.chainID,
		NetworkShardingCollector: n.networkShardingCollector,
		AntifloodHandler:         n.inputAntifloodHandler,
		PoolAdder:                n.dataPool.MiniBlocks(),
		SignatureSize:            n.signatureSize,
		PublicKeySize:            n.publicKeySize,
	}

	worker, err := spos.NewWorker(workerArgs)
	if err != nil {
		return err
	}

	worker.StartWorking()

	n.dataPool.Headers().RegisterHandler(worker.ReceivedHeader)

	// apply consensus group size on the input antiflooder just befor consensus creation topic
	n.inputAntifloodHandler.ApplyConsensusSize(n.nodesCoordinator.ConsensusGroupSize(n.shardCoordinator.SelfId()))
	err = n.createConsensusTopic(worker)
	if err != nil {
		return err
	}

	consensusArgs := &spos.ConsensusCoreArgs{
		BlockChain:                    n.blkc,
		BlockProcessor:                n.blockProcessor,
		Bootstrapper:                  bootstrapper,
		BroadcastMessenger:            broadcastMessenger,
		ChronologyHandler:             chronologyHandler,
		Hasher:                        n.hasher,
		Marshalizer:                   n.internalMarshalizer,
		BlsPrivateKey:                 n.privKey,
		BlsSingleSigner:               n.singleSigner,
		MultiSigner:                   n.multiSigner,
		Rounder:                       n.rounder,
		ShardCoordinator:              n.shardCoordinator,
		NodesCoordinator:              n.nodesCoordinator,
		SyncTimer:                     n.syncTimer,
		EpochStartRegistrationHandler: n.epochStartRegistrationHandler,
		AntifloodHandler:              n.inputAntifloodHandler,
		PeerHonestyHandler:            n.peerHonestyHandler,
	}

	consensusDataContainer, err := spos.NewConsensusCore(
		consensusArgs,
	)
	if err != nil {
		return err
	}

	fct, err := sposFactory.GetSubroundsFactory(
		consensusDataContainer,
		consensusState,
		worker,
		n.consensusType,
		n.appStatusHandler,
		n.indexer,
		n.chainID,
		n.messenger.ID(),
	)
	if err != nil {
		return err
	}

	err = fct.GenerateSubrounds()
	if err != nil {
		return err
	}

	chronologyHandler.StartRounds()

	return n.addCloserInstances(chronologyHandler, bootstrapper, worker, n.syncTimer)
}

func (n *Node) addCloserInstances(closers ...update.Closer) error {
	for _, c := range closers {
		err := n.hardforkTrigger.AddCloser(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetBalance gets the balance for a specific address
func (n *Node) GetBalance(address string) (*big.Int, error) {
	if check.IfNil(n.addressPubkeyConverter) || check.IfNil(n.accounts) {
		return nil, errors.New("initialize AccountsAdapter and PubkeyConverter first")
	}

	addr, err := n.addressPubkeyConverter.Decode(address)
	if err != nil {
		return nil, errors.New("invalid address, could not decode from: " + err.Error())
	}
	accWrp, err := n.accounts.GetExistingAccount(addr)
	if err != nil {
		return nil, errors.New("could not fetch sender address from provided param: " + err.Error())
	}

	if check.IfNil(accWrp) {
		return big.NewInt(0), nil
	}

	account, ok := accWrp.(state.UserAccountHandler)
	if !ok {
		return big.NewInt(0), nil
	}

	return account.GetBalance(), nil
}

// GetValueForKey will return the value for a key from a given account
func (n *Node) GetValueForKey(address string, key string) (string, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("invalid key: %w", err)
	}

	if check.IfNil(n.addressPubkeyConverter) || check.IfNil(n.accounts) {
		return "", fmt.Errorf("initialize AccountsAdapter and PubkeyConverter first")
	}

	addr, err := n.addressPubkeyConverter.Decode(address)
	if err != nil {
		return "", fmt.Errorf("invalid address, could not decode from: %w", err)
	}
	accWrp, err := n.accounts.GetExistingAccount(addr)
	if err != nil {
		return "", fmt.Errorf("could not fetch sender address from provided param: %w", err)
	}

	if check.IfNil(accWrp) {
		return "", fmt.Errorf("account not found")
	}
	account, ok := accWrp.(state.UserAccountHandler)
	if !ok {
		return "", fmt.Errorf("account not found - cannot convert to UserAccountHandler")
	}

	valueBytes, err := account.DataTrieTracker().RetrieveValue(keyBytes)
	if err != nil {
		return "", fmt.Errorf("fetching value error: %w", err)
	}

	return hex.EncodeToString(valueBytes), nil
}

// createChronologyHandler method creates a chronology object
func (n *Node) createChronologyHandler(
	rounder consensus.Rounder,
	appStatusHandler core.AppStatusHandler,
	watchdog core.WatchdogTimer,
) (consensus.ChronologyHandler, error) {
	chr, err := chronology.NewChronology(
		n.genesisTime,
		rounder,
		n.syncTimer,
		watchdog,
	)
	if err != nil {
		return nil, err
	}

	err = chr.SetAppStatusHandler(appStatusHandler)
	if err != nil {
		return nil, err
	}

	return chr, nil
}

// TODO move this func in structs.go
func (n *Node) createBootstrapper(rounder consensus.Rounder) (process.Bootstrapper, error) {
	miniblocksProvider, err := n.createMiniblocksProvider()
	if err != nil {
		return nil, err
	}

	n.miniblocksProvider = miniblocksProvider

	if n.shardCoordinator.SelfId() < n.shardCoordinator.NumberOfShards() {
		return n.createShardBootstrapper(rounder)
	}

	if n.shardCoordinator.SelfId() == core.MetachainShardId {
		return n.createMetaChainBootstrapper(rounder)
	}

	return nil, sharding.ErrShardIdOutOfRange
}

func (n *Node) createShardBootstrapper(rounder consensus.Rounder) (process.Bootstrapper, error) {
	argsBaseStorageBootstrapper := storageBootstrap.ArgsBaseStorageBootstrapper{
		BootStorer:          n.bootStorer,
		ForkDetector:        n.forkDetector,
		BlockProcessor:      n.blockProcessor,
		ChainHandler:        n.blkc,
		Marshalizer:         n.internalMarshalizer,
		Store:               n.store,
		Uint64Converter:     n.uint64ByteSliceConverter,
		BootstrapRoundIndex: n.bootstrapRoundIndex,
		ShardCoordinator:    n.shardCoordinator,
		NodesCoordinator:    n.nodesCoordinator,
		EpochStartTrigger:   n.epochStartTrigger,
		BlockTracker:        n.blockTracker,
		ChainID:             string(n.chainID),
	}

	argsShardStorageBootstrapper := storageBootstrap.ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: argsBaseStorageBootstrapper,
	}

	shardStorageBootstrapper, err := storageBootstrap.NewShardStorageBootstrapper(argsShardStorageBootstrapper)
	if err != nil {
		return nil, err
	}

	argsBaseBootstrapper := sync.ArgBaseBootstrapper{
		PoolsHolder:         n.dataPool,
		Store:               n.store,
		ChainHandler:        n.blkc,
		Rounder:             rounder,
		BlockProcessor:      n.blockProcessor,
		WaitTime:            n.rounder.TimeDuration(),
		Hasher:              n.hasher,
		Marshalizer:         n.internalMarshalizer,
		ForkDetector:        n.forkDetector,
		RequestHandler:      n.requestHandler,
		ShardCoordinator:    n.shardCoordinator,
		Accounts:            n.accounts,
		BlackListHandler:    n.blocksBlackListHandler,
		NetworkWatcher:      n.messenger,
		BootStorer:          n.bootStorer,
		StorageBootstrapper: shardStorageBootstrapper,
		EpochHandler:        n.epochStartTrigger,
		MiniblocksProvider:  n.miniblocksProvider,
		Uint64Converter:     n.uint64ByteSliceConverter,
	}

	argsShardBootstrapper := sync.ArgShardBootstrapper{
		ArgBaseBootstrapper: argsBaseBootstrapper,
	}

	bootstrap, err := sync.NewShardBootstrap(argsShardBootstrapper)
	if err != nil {
		return nil, err
	}

	return bootstrap, nil
}

func (n *Node) createMetaChainBootstrapper(rounder consensus.Rounder) (process.Bootstrapper, error) {
	argsBaseStorageBootstrapper := storageBootstrap.ArgsBaseStorageBootstrapper{
		BootStorer:          n.bootStorer,
		ForkDetector:        n.forkDetector,
		BlockProcessor:      n.blockProcessor,
		ChainHandler:        n.blkc,
		Marshalizer:         n.internalMarshalizer,
		Store:               n.store,
		Uint64Converter:     n.uint64ByteSliceConverter,
		BootstrapRoundIndex: n.bootstrapRoundIndex,
		ShardCoordinator:    n.shardCoordinator,
		NodesCoordinator:    n.nodesCoordinator,
		EpochStartTrigger:   n.epochStartTrigger,
		BlockTracker:        n.blockTracker,
		ChainID:             string(n.chainID),
	}

	argsMetaStorageBootstrapper := storageBootstrap.ArgsMetaStorageBootstrapper{
		ArgsBaseStorageBootstrapper: argsBaseStorageBootstrapper,
		PendingMiniBlocksHandler:    n.pendingMiniBlocksHandler,
	}

	metaStorageBootstrapper, err := storageBootstrap.NewMetaStorageBootstrapper(argsMetaStorageBootstrapper)
	if err != nil {
		return nil, err
	}

	argsBaseBootstrapper := sync.ArgBaseBootstrapper{
		PoolsHolder:         n.dataPool,
		Store:               n.store,
		ChainHandler:        n.blkc,
		Rounder:             rounder,
		BlockProcessor:      n.blockProcessor,
		WaitTime:            n.rounder.TimeDuration(),
		Hasher:              n.hasher,
		Marshalizer:         n.internalMarshalizer,
		ForkDetector:        n.forkDetector,
		RequestHandler:      n.requestHandler,
		ShardCoordinator:    n.shardCoordinator,
		Accounts:            n.accounts,
		BlackListHandler:    n.blocksBlackListHandler,
		NetworkWatcher:      n.messenger,
		BootStorer:          n.bootStorer,
		StorageBootstrapper: metaStorageBootstrapper,
		EpochHandler:        n.epochStartTrigger,
		MiniblocksProvider:  n.miniblocksProvider,
		Uint64Converter:     n.uint64ByteSliceConverter,
	}

	argsMetaBootstrapper := sync.ArgMetaBootstrapper{
		ArgBaseBootstrapper: argsBaseBootstrapper,
		EpochBootstrapper:   n.epochStartTrigger,
	}

	bootstrap, err := sync.NewMetaBootstrap(argsMetaBootstrapper)
	if err != nil {
		return nil, err
	}

	return bootstrap, nil
}

func (n *Node) createMiniblocksProvider() (process.MiniBlockProvider, error) {
	if check.IfNil(n.dataPool) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(n.store) {
		return nil, process.ErrNilStorage
	}

	arg := provider.ArgMiniBlockProvider{
		MiniBlockPool:    n.dataPool.MiniBlocks(),
		MiniBlockStorage: n.store.GetStorer(dataRetriever.MiniBlockUnit),
		Marshalizer:      n.internalMarshalizer,
	}

	return provider.NewMiniBlockProvider(arg)
}

// createConsensusState method creates a consensusState object
func (n *Node) createConsensusState(epoch uint32) (*spos.ConsensusState, error) {
	selfId, err := n.pubKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	eligibleNodesPubKeys, err := n.nodesCoordinator.GetConsensusWhitelistedNodes(epoch)
	if err != nil {
		return nil, err
	}

	roundConsensus := spos.NewRoundConsensus(
		eligibleNodesPubKeys,
		n.consensusGroupSize,
		string(selfId))

	roundConsensus.ResetRoundState()

	roundThreshold := spos.NewRoundThreshold()

	roundStatus := spos.NewRoundStatus()
	roundStatus.ResetRoundStatus()

	consensusState := spos.NewConsensusState(
		roundConsensus,
		roundThreshold,
		roundStatus)

	return consensusState, nil
}

// createConsensusTopic creates a consensus topic for node
func (n *Node) createConsensusTopic(messageProcessor p2p.MessageProcessor) error {
	if check.IfNil(n.shardCoordinator) {
		return ErrNilShardCoordinator
	}
	if check.IfNil(messageProcessor) {
		return ErrNilMessenger
	}

	n.consensusTopic = core.ConsensusTopic + n.shardCoordinator.CommunicationIdentifier(n.shardCoordinator.SelfId())
	if !n.messenger.HasTopic(n.consensusTopic) {
		err := n.messenger.CreateTopic(n.consensusTopic, true)
		if err != nil {
			return err
		}
	}

	if n.messenger.HasTopicValidator(n.consensusTopic) {
		return ErrValidatorAlreadySet
	}

	return n.messenger.RegisterMessageProcessor(n.consensusTopic, messageProcessor)
}

// SendBulkTransactions sends the provided transactions as a bulk, optimizing transfer between nodes
func (n *Node) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	if len(txs) == 0 {
		return 0, ErrNoTxToProcess
	}

	n.addTransactionsToSendPipe(txs)

	return uint64(len(txs)), nil
}

func (n *Node) addTransactionsToSendPipe(txs []*transaction.Transaction) {
	if check.IfNil(n.txAcumulator) {
		log.Error("node has a nil tx accumulator instance")
		return
	}

	for _, tx := range txs {
		n.txAcumulator.AddData(tx)
	}
}

func (n *Node) sendFromTxAccumulator() {
	outputChannel := n.txAcumulator.OutputChannel()

	for objs := range outputChannel {
		//this will read continuously until the chan is closed

		if len(objs) == 0 {
			continue
		}

		txs := make([]*transaction.Transaction, 0, len(objs))
		for _, obj := range objs {
			tx, ok := obj.(*transaction.Transaction)
			if !ok {
				continue
			}

			txs = append(txs, tx)
		}

		atomic.AddUint32(&n.txSentCounter, uint32(len(txs)))

		n.sendBulkTransactions(txs)
	}
}

// printTxSentCounter prints the peak transaction counter from a time frame of about 'numSecondsBetweenPrints' seconds
// if this peak value is 0 (no transaction was sent through the REST API interface), the print will not be done
// the peak counter resets after each print. There is also a total number of transactions sent to p2p
// TODO make this function testable. Refactor if necessary.
func (n *Node) printTxSentCounter() {
	maxTxCounter := uint32(0)
	totalTxCounter := uint64(0)
	counterSeconds := 0

	for {
		time.Sleep(time.Second)

		txSent := atomic.SwapUint32(&n.txSentCounter, 0)
		if txSent > maxTxCounter {
			maxTxCounter = txSent
		}
		totalTxCounter += uint64(txSent)

		counterSeconds++
		if counterSeconds > numSecondsBetweenPrints {
			counterSeconds = 0

			if maxTxCounter > 0 {
				log.Info("sent transactions on network",
					"max/sec", maxTxCounter,
					"total", totalTxCounter,
				)
			}
			maxTxCounter = 0
		}
	}
}

// sendBulkTransactions sends the provided transactions as a bulk, optimizing transfer between nodes
func (n *Node) sendBulkTransactions(txs []*transaction.Transaction) {
	transactionsByShards := make(map[uint32][][]byte)
	log.Trace("node.sendBulkTransactions sending txs",
		"num", len(txs),
	)

	for _, tx := range txs {
		senderShardId := n.shardCoordinator.ComputeId(tx.SndAddr)

		marshalizedTx, err := n.internalMarshalizer.Marshal(tx)
		if err != nil {
			log.Warn("node.sendBulkTransactions",
				"marshalizer error", err,
			)
			continue
		}

		transactionsByShards[senderShardId] = append(transactionsByShards[senderShardId], marshalizedTx)
	}

	numOfSentTxs := uint64(0)
	for shardId, txsForShard := range transactionsByShards {
		err := n.sendBulkTransactionsFromShard(txsForShard, shardId)
		if err != nil {
			log.Debug("sendBulkTransactionsFromShard", "error", err.Error())
		} else {
			numOfSentTxs += uint64(len(txsForShard))
		}
	}
}

// ValidateTransaction will validate a transaction
func (n *Node) ValidateTransaction(tx *transaction.Transaction) error {
	txValidator, err := dataValidators.NewTxValidator(
		n.accounts,
		n.shardCoordinator,
		n.whiteListRequest,
		n.addressPubkeyConverter,
		core.MaxTxNonceDeltaAllowed,
	)
	if err != nil {
		return nil
	}

	marshalizedTx, err := n.internalMarshalizer.Marshal(tx)
	if err != nil {
		return err
	}

	argumentParser := smartContract.NewArgumentParser()
	intTx, err := procTx.NewInterceptedTransaction(
		marshalizedTx,
		n.internalMarshalizer,
		n.txSignMarshalizer,
		n.hasher,
		n.keyGenForAccounts,
		n.txSingleSigner,
		n.addressPubkeyConverter,
		n.shardCoordinator,
		n.feeHandler,
		n.whiteListerVerifiedTxs,
		argumentParser,
		n.chainID,
		n.minTransactionVersion,
	)
	if err != nil {
		return err
	}

	err = intTx.CheckValidity()
	if err != nil {
		return err
	}

	err = txValidator.CheckTxValidity(intTx)
	if errors.Is(err, process.ErrAccountNotFound) {
		// we allow the broadcast of provided transaction even if that transaction is not targeted on the current shard
		return nil
	}

	return err
}

func (n *Node) sendBulkTransactionsFromShard(transactions [][]byte, senderShardId uint32) error {
	dataPacker, err := partitioning.NewSimpleDataPacker(n.internalMarshalizer)
	if err != nil {
		return err
	}

	//the topic identifier is made of the current shard id and sender's shard id
	identifier := factory.TransactionTopic + n.shardCoordinator.CommunicationIdentifier(senderShardId)

	packets, err := dataPacker.PackDataInChunks(transactions, core.MaxBulkTransactionSize)
	if err != nil {
		return err
	}

	atomic.AddInt32(&n.currentSendingGoRoutines, int32(len(packets)))
	for _, buff := range packets {
		go func(bufferToSend []byte) {
			log.Trace("node.sendBulkTransactionsFromShard",
				"topic", identifier,
				"size", len(bufferToSend),
			)
			err = n.messenger.BroadcastOnChannelBlocking(
				SendTransactionsPipe,
				identifier,
				bufferToSend,
			)
			if err != nil {
				log.Debug("node.BroadcastOnChannelBlocking", "error", err.Error())
			}

			atomic.AddInt32(&n.currentSendingGoRoutines, -1)
		}(buff)
	}

	return nil
}

// CreateTransaction will return a transaction from all the required fields
func (n *Node) CreateTransaction(
	nonce uint64,
	value string,
	receiver string,
	sender string,
	gasPrice uint64,
	gasLimit uint64,
	dataField []byte,
	signatureHex string,
	chainID string,
	version uint32,
) (*transaction.Transaction, []byte, error) {
	if version == 0 {
		return nil, nil, ErrInvalidTransactionVersion
	}
	if chainID == "" {
		return nil, nil, ErrInvalidChainID
	}
	if check.IfNil(n.addressPubkeyConverter) {
		return nil, nil, ErrNilPubkeyConverter
	}
	if check.IfNil(n.accounts) {
		return nil, nil, ErrNilAccountsAdapter
	}

	receiverAddress, err := n.addressPubkeyConverter.Decode(receiver)
	if err != nil {
		return nil, nil, errors.New("could not create receiver address from provided param")
	}

	senderAddress, err := n.addressPubkeyConverter.Decode(sender)
	if err != nil {
		return nil, nil, errors.New("could not create sender address from provided param")
	}

	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return nil, nil, errors.New("could not fetch signature bytes")
	}

	valAsBigInt, ok := big.NewInt(0).SetString(value, 10)
	if !ok {
		return nil, nil, ErrInvalidValue
	}

	tx := &transaction.Transaction{
		Nonce:     nonce,
		Value:     valAsBigInt,
		RcvAddr:   receiverAddress,
		SndAddr:   senderAddress,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      dataField,
		Signature: signatureBytes,
		ChainID:   []byte(chainID),
		Version:   version,
	}

	var txHash []byte
	txHash, err = core.CalculateHash(n.internalMarshalizer, n.hasher, tx)
	if err != nil {
		return nil, nil, err
	}

	return tx, txHash, nil
}

// GetAccount will return account details for a given address
func (n *Node) GetAccount(address string) (state.UserAccountHandler, error) {
	if check.IfNil(n.addressPubkeyConverter) {
		return nil, ErrNilPubkeyConverter
	}
	if check.IfNil(n.accounts) {
		return nil, ErrNilAccountsAdapter
	}

	addr, err := n.addressPubkeyConverter.Decode(address)
	if err != nil {
		return nil, err
	}

	accWrp, err := n.accounts.GetExistingAccount(addr)
	if err != nil {
		if err == state.ErrAccNotFound {
			return state.NewUserAccount(addr)
		}
		return nil, errors.New("could not fetch sender address from provided param: " + err.Error())
	}

	account, ok := accWrp.(state.UserAccountHandler)
	if !ok {
		return nil, errors.New("account is not of type with balance and nonce")
	}

	return account, nil
}

// StartHeartbeat starts the node's heartbeat processing/signaling module
// TODO(next PR) remove the instantiation of the heartbeat component from here
func (n *Node) StartHeartbeat(hbConfig config.HeartbeatConfig, versionNumber string, prefsConfig config.PreferencesConfig) error {
	arg := componentHandler.ArgHeartbeat{
		HeartbeatConfig:          hbConfig,
		PrefsConfig:              prefsConfig,
		Marshalizer:              n.internalMarshalizer,
		Messenger:                n.messenger,
		ShardCoordinator:         n.shardCoordinator,
		NodesCoordinator:         n.nodesCoordinator,
		AppStatusHandler:         n.appStatusHandler,
		Storer:                   n.store.GetStorer(dataRetriever.HeartbeatUnit),
		ValidatorStatistics:      n.validatorStatistics,
		PeerSignatureHandler:     n.peerSigHandler,
		PrivKey:                  n.privKey,
		HardforkTrigger:          n.hardforkTrigger,
		AntifloodHandler:         n.inputAntifloodHandler,
		ValidatorPubkeyConverter: n.validatorPubkeyConverter,
		EpochStartTrigger:        n.epochStartTrigger,
		EpochStartRegistration:   n.epochStartRegistrationHandler,
		Timer:                    &heartbeatProcess.RealTimer{},
		GenesisTime:              n.genesisTime,
		VersionNumber:            versionNumber,
		PeerShardMapper:          n.networkShardingCollector,
		SizeCheckDelta:           n.sizeCheckDelta,
		ValidatorsProvider:       n.validatorsProvider,
	}

	var err error
	n.heartbeatHandler, err = componentHandler.NewHeartbeatHandler(arg)

	return err
}

// GetHeartbeats returns the heartbeat status for each public key defined in genesis.json
func (n *Node) GetHeartbeats() []heartbeatData.PubKeyHeartbeat {
	if check.IfNil(n.heartbeatHandler) {
		return make([]heartbeatData.PubKeyHeartbeat, 0)
	}
	mon := n.heartbeatHandler.Monitor()
	if check.IfNil(mon) {
		return make([]heartbeatData.PubKeyHeartbeat, 0)
	}

	return mon.GetHeartbeats()
}

// ValidatorStatisticsApi will return the statistics for all the validators from the initial nodes pub keys
func (n *Node) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return n.validatorsProvider.GetLatestValidators(), nil
}

func (n *Node) getLatestValidators() (map[uint32][]*state.ValidatorInfo, map[string]*state.ValidatorApiResponse, error) {
	latestHash, err := n.validatorStatistics.RootHash()
	if err != nil {
		return nil, nil, err
	}

	validators, err := n.validatorStatistics.GetValidatorInfoForRootHash(latestHash)
	if err != nil {
		return nil, nil, err
	}

	return validators, nil, nil
}

// DirectTrigger will start the hardfork trigger
func (n *Node) DirectTrigger(epoch uint32) error {
	return n.hardforkTrigger.Trigger(epoch)
}

// IsSelfTrigger returns true if the trigger's registered public key matches the self public key
func (n *Node) IsSelfTrigger() bool {
	return n.hardforkTrigger.IsSelfTrigger()
}

// EncodeAddressPubkey will encode the provided address public key bytes to string
func (n *Node) EncodeAddressPubkey(pk []byte) (string, error) {
	if n.addressPubkeyConverter == nil {
		return "", fmt.Errorf("%w for addressPubkeyConverter", ErrNilPubkeyConverter)
	}

	return n.addressPubkeyConverter.Encode(pk), nil
}

// DecodeAddressPubkey will try to decode the provided address public key string
func (n *Node) DecodeAddressPubkey(pk string) ([]byte, error) {
	if n.addressPubkeyConverter == nil {
		return nil, fmt.Errorf("%w for addressPubkeyConverter", ErrNilPubkeyConverter)
	}

	return n.addressPubkeyConverter.Decode(pk)
}

// AddQueryHandler adds a query handler in cache
func (n *Node) AddQueryHandler(name string, handler debug.QueryHandler) error {
	if check.IfNil(handler) {
		return ErrNilQueryHandler
	}
	if len(name) == 0 {
		return ErrEmptyQueryHandlerName
	}

	n.mutQueryHandlers.Lock()
	defer n.mutQueryHandlers.Unlock()

	_, ok := n.queryHandlers[name]
	if ok {
		return fmt.Errorf("%w with name %s", ErrQueryHandlerAlreadyExists, name)
	}

	n.queryHandlers[name] = handler

	return nil
}

// GetQueryHandler returns the query handler if existing
func (n *Node) GetQueryHandler(name string) (debug.QueryHandler, error) {
	n.mutQueryHandlers.RLock()
	defer n.mutQueryHandlers.RUnlock()

	qh, ok := n.queryHandlers[name]
	if !ok || check.IfNil(qh) {
		return nil, fmt.Errorf("%w for name %s", ErrNilQueryHandler, name)
	}

	return qh, nil
}

// GetPeerInfo returns information about a peer id
func (n *Node) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	peers := n.messenger.Peers()
	pidsFound := make([]core.PeerID, 0)
	for _, p := range peers {
		if strings.Contains(p.Pretty(), pid) {
			pidsFound = append(pidsFound, p)
		}
	}

	if len(pidsFound) == 0 {
		return nil, fmt.Errorf("%w for provided peer %s", ErrUnknownPeerID, pid)
	}

	sort.Slice(pidsFound, func(i, j int) bool {
		return pidsFound[i].Pretty() < pidsFound[j].Pretty()
	})

	peerInfoSlice := make([]core.QueryP2PPeerInfo, 0, len(pidsFound))
	for _, p := range pidsFound {
		pidInfo := n.createPidInfo(p)
		peerInfoSlice = append(peerInfoSlice, pidInfo)
	}

	return peerInfoSlice, nil
}

func (n *Node) createPidInfo(p core.PeerID) core.QueryP2PPeerInfo {
	result := core.QueryP2PPeerInfo{
		Pid:           p.Pretty(),
		Addresses:     n.messenger.PeerAddresses(p),
		IsBlacklisted: n.peerDenialEvaluator.IsDenied(p),
	}

	peerInfo := n.networkShardingCollector.GetPeerInfo(p)
	result.PeerType = peerInfo.PeerType.String()
	if len(peerInfo.PkBytes) == 0 {
		result.Pk = ""
	} else {
		result.Pk = n.validatorPubkeyConverter.Encode(peerInfo.PkBytes)
	}

	return result
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *Node) IsInterfaceNil() bool {
	return n == nil
}
