package coordinator

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/batch"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/block/preprocess"
	"github.com/Dharitri-org/sme-dharitri/process/factory"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/storage"
	"github.com/Dharitri-org/sme-dharitri/storage/timecache"
	logger "github.com/Dharitri-org/sme-logger"
)

var _ process.TransactionCoordinator = (*transactionCoordinator)(nil)

var log = logger.GetOrCreate("process/coordinator")

type transactionCoordinator struct {
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter
	miniBlockPool    storage.Cacher
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer

	mutPreProcessor sync.RWMutex
	txPreProcessors map[block.Type]process.PreProcessor
	keysTxPreProcs  []block.Type

	mutInterimProcessors sync.RWMutex
	interimProcessors    map[block.Type]process.IntermediateTransactionHandler
	keysInterimProcs     []block.Type

	mutRequestedTxs sync.RWMutex
	requestedTxs    map[block.Type]int

	onRequestMiniBlock    func(shardId uint32, mbHash []byte)
	gasHandler            process.GasHandler
	feeHandler            process.TransactionFeeHandler
	blockSizeComputation  preprocess.BlockSizeComputationHandler
	balanceComputation    preprocess.BalanceComputationHandler
	requestedItemsHandler process.TimeCacher
}

// NewTransactionCoordinator creates a transaction coordinator to run and coordinate preprocessors and processors
func NewTransactionCoordinator(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	miniBlockPool storage.Cacher,
	requestHandler process.RequestHandler,
	preProcessors process.PreProcessorsContainer,
	interProcessors process.IntermediateProcessorContainer,
	gasHandler process.GasHandler,
	feeHandler process.TransactionFeeHandler,
	blockSizeComputation preprocess.BlockSizeComputationHandler,
	balanceComputation preprocess.BalanceComputationHandler,
) (*transactionCoordinator, error) {

	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(miniBlockPool) {
		return nil, process.ErrNilMiniBlockPool
	}
	if check.IfNil(requestHandler) {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(interProcessors) {
		return nil, process.ErrNilIntermediateProcessorContainer
	}
	if check.IfNil(preProcessors) {
		return nil, process.ErrNilPreProcessorsContainer
	}
	if check.IfNil(gasHandler) {
		return nil, process.ErrNilGasHandler
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(feeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(blockSizeComputation) {
		return nil, process.ErrNilBlockSizeComputationHandler
	}
	if check.IfNil(balanceComputation) {
		return nil, process.ErrNilBalanceComputationHandler
	}

	tc := &transactionCoordinator{
		shardCoordinator:     shardCoordinator,
		accounts:             accounts,
		gasHandler:           gasHandler,
		hasher:               hasher,
		marshalizer:          marshalizer,
		feeHandler:           feeHandler,
		blockSizeComputation: blockSizeComputation,
		balanceComputation:   balanceComputation,
	}

	tc.miniBlockPool = miniBlockPool
	tc.onRequestMiniBlock = requestHandler.RequestMiniBlock
	tc.requestedTxs = make(map[block.Type]int)
	tc.txPreProcessors = make(map[block.Type]process.PreProcessor)
	tc.interimProcessors = make(map[block.Type]process.IntermediateTransactionHandler)

	tc.keysTxPreProcs = preProcessors.Keys()
	sort.Slice(tc.keysTxPreProcs, func(i, j int) bool {
		return tc.keysTxPreProcs[i] < tc.keysTxPreProcs[j]
	})
	for _, value := range tc.keysTxPreProcs {
		preProc, err := preProcessors.Get(value)
		if err != nil {
			return nil, err
		}
		tc.txPreProcessors[value] = preProc
	}

	tc.keysInterimProcs = interProcessors.Keys()
	sort.Slice(tc.keysInterimProcs, func(i, j int) bool {
		return tc.keysInterimProcs[i] < tc.keysInterimProcs[j]
	})
	for _, value := range tc.keysInterimProcs {
		interProc, err := interProcessors.Get(value)
		if err != nil {
			return nil, err
		}
		tc.interimProcessors[value] = interProc
	}

	tc.requestedItemsHandler = timecache.NewTimeCache(core.MaxWaitingTimeToReceiveRequestedItem)
	tc.miniBlockPool.RegisterHandler(tc.receivedMiniBlock, core.UniqueIdentifier())

	return tc, nil
}

// separateBodyByType creates a map of bodies according to type
func (tc *transactionCoordinator) separateBodyByType(body *block.Body) map[block.Type]*block.Body {
	separatedBodies := make(map[block.Type]*block.Body)
	for i := 0; i < len(body.MiniBlocks); i++ {
		mb := body.MiniBlocks[i]

		separatedMbType := mb.Type
		if mb.Type == block.InvalidBlock {
			separatedMbType = block.TxBlock
		}

		if _, ok := separatedBodies[separatedMbType]; !ok {
			separatedBodies[separatedMbType] = &block.Body{}
		}

		separatedBodies[separatedMbType].MiniBlocks = append(separatedBodies[separatedMbType].MiniBlocks, mb)
	}

	return separatedBodies
}

// initRequestedTxs init the requested txs number
func (tc *transactionCoordinator) initRequestedTxs() {
	tc.mutRequestedTxs.Lock()
	tc.requestedTxs = make(map[block.Type]int)
	tc.mutRequestedTxs.Unlock()
}

// RequestBlockTransactions verifies missing transaction and requests them
func (tc *transactionCoordinator) RequestBlockTransactions(body *block.Body) {
	if check.IfNil(body) {
		return
	}

	separatedBodies := tc.separateBodyByType(body)

	tc.initRequestedTxs()

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody *block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}
			requestedTxs := preproc.RequestBlockTransactions(blockBody)

			tc.mutRequestedTxs.Lock()
			tc.requestedTxs[blockType] = requestedTxs
			tc.mutRequestedTxs.Unlock()

			wg.Done()
		}(key, value)
	}

	wg.Wait()
}

// IsDataPreparedForProcessing verifies if all the needed data is prepared
func (tc *transactionCoordinator) IsDataPreparedForProcessing(haveTime func() time.Duration) error {
	var errFound error
	errMutex := sync.Mutex{}

	wg := sync.WaitGroup{}

	tc.mutRequestedTxs.RLock()
	wg.Add(len(tc.requestedTxs))

	for key, value := range tc.requestedTxs {
		go func(blockType block.Type, requestedTxs int) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}

			err := preproc.IsDataPrepared(requestedTxs, haveTime)
			if err != nil {
				log.Trace("IsDataPrepared", "error", err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}(key, value)
	}

	tc.mutRequestedTxs.RUnlock()
	wg.Wait()

	return errFound
}

// SaveBlockDataToStorage saves the data from block body into storage units
func (tc *transactionCoordinator) SaveBlockDataToStorage(body *block.Body) error {
	if check.IfNil(body) {
		return nil
	}

	separatedBodies := tc.separateBodyByType(body)
	for key, value := range separatedBodies {
		err := tc.saveTxBlockToStorage(key, value)
		if err != nil {
			return err
		}
	}

	for _, blockType := range tc.keysInterimProcs {
		err := tc.saveCurrentIntermediateTxToStorage(blockType)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tc *transactionCoordinator) saveTxBlockToStorage(blockType block.Type, blockBody *block.Body) error {
	preproc := tc.getPreProcessor(blockType)
	if check.IfNil(preproc) {
		return nil
	}

	err := preproc.SaveTxBlockToStorage(blockBody)
	if err != nil {
		log.Trace("SaveTxBlockToStorage", "error", err.Error())

		return err
	}

	return nil
}

func (tc *transactionCoordinator) saveCurrentIntermediateTxToStorage(blockType block.Type) error {
	intermediateProc := tc.getInterimProcessor(blockType)
	if check.IfNil(intermediateProc) {
		return nil
	}

	err := intermediateProc.SaveCurrentIntermediateTxToStorage()
	if err != nil {
		log.Trace("SaveCurrentIntermediateTxToStorage", "error", err.Error())
		return err
	}

	return nil
}

// RestoreBlockDataFromStorage restores block data from storage to pool
func (tc *transactionCoordinator) RestoreBlockDataFromStorage(body *block.Body) (int, error) {
	if check.IfNil(body) {
		return 0, nil
	}

	separatedBodies := tc.separateBodyByType(body)

	var errFound error
	localMutex := sync.Mutex{}
	totalRestoredTx := 0

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody *block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}

			restoredTxs, err := preproc.RestoreTxBlockIntoPools(blockBody, tc.miniBlockPool)
			if err != nil {
				log.Trace("RestoreTxBlockIntoPools", "error", err.Error())

				localMutex.Lock()
				errFound = err
				localMutex.Unlock()
			}

			localMutex.Lock()
			totalRestoredTx += restoredTxs

			localMutex.Unlock()

			wg.Done()
		}(key, value)
	}

	wg.Wait()

	return totalRestoredTx, errFound
}

// RemoveBlockDataFromPool deletes block data from pools
func (tc *transactionCoordinator) RemoveBlockDataFromPool(body *block.Body) error {
	if check.IfNil(body) {
		return nil
	}

	separatedBodies := tc.separateBodyByType(body)

	var errFound error
	errMutex := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody *block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}

			err := preproc.RemoveTxBlockFromPools(blockBody, tc.miniBlockPool)
			if err != nil {
				log.Trace("RemoveTxBlockFromPools", "error", err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}(key, value)
	}

	wg.Wait()

	return errFound
}

// ProcessBlockTransaction processes transactions and updates state tries
func (tc *transactionCoordinator) ProcessBlockTransaction(
	body *block.Body,
	timeRemaining func() time.Duration,
) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}

	haveTime := func() bool {
		return timeRemaining() >= 0
	}

	for _, miniBlock := range body.MiniBlocks {
		log.Trace("ProcessBlockTransaction: miniblock",
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"type", miniBlock.Type,
			"num txs", len(miniBlock.TxHashes))
	}

	startTime := time.Now()
	mbIndex, err := tc.processMiniBlocksToMe(body, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to processMiniBlocksToMe",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	if mbIndex == len(body.MiniBlocks) {
		return nil
	}

	miniBlocksFromMe := body.MiniBlocks[mbIndex:]
	startTime = time.Now()
	err = tc.processMiniBlocksFromMe(&block.Body{MiniBlocks: miniBlocksFromMe}, haveTime)
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to processMiniBlocksFromMe",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	return nil
}

func (tc *transactionCoordinator) processMiniBlocksFromMe(
	body *block.Body,
	haveTime func() bool,
) error {
	for _, mb := range body.MiniBlocks {
		if mb.SenderShardID != tc.shardCoordinator.SelfId() {
			return process.ErrMiniBlocksInWrongOrder
		}
	}

	separatedBodies := tc.separateBodyByType(body)
	// processing has to be done in order, as the order of different type of transactions over the same account is strict
	for _, blockType := range tc.keysTxPreProcs {
		if separatedBodies[blockType] == nil {
			continue
		}

		preProc := tc.getPreProcessor(blockType)
		if check.IfNil(preProc) {
			return process.ErrMissingPreProcessor
		}

		err := preProc.ProcessBlockTransactions(separatedBodies[blockType], haveTime)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tc *transactionCoordinator) processMiniBlocksToMe(
	body *block.Body,
	haveTime func() bool,
) (int, error) {
	// processing has to be done in order, as the order of different type of transactions over the same account is strict
	// processing destination ME miniblocks first
	mbIndex := 0
	for mbIndex = 0; mbIndex < len(body.MiniBlocks); mbIndex++ {
		miniBlock := body.MiniBlocks[mbIndex]
		if miniBlock.SenderShardID == tc.shardCoordinator.SelfId() {
			return mbIndex, nil
		}

		preProc := tc.getPreProcessor(miniBlock.Type)
		if check.IfNil(preProc) {
			return mbIndex, process.ErrMissingPreProcessor
		}

		err := preProc.ProcessBlockTransactions(&block.Body{MiniBlocks: []*block.MiniBlock{miniBlock}}, haveTime)
		if err != nil {
			return mbIndex, err
		}
	}

	return mbIndex, nil
}

// CreateMbsAndProcessCrossShardTransactionsDstMe creates miniblocks and processes cross shard transaction
// with destination of current shard
func (tc *transactionCoordinator) CreateMbsAndProcessCrossShardTransactionsDstMe(
	hdr data.HeaderHandler,
	processedMiniBlocksHashes map[string]struct{},
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, bool, error) {

	miniBlocks := make(block.MiniBlockSlice, 0)
	nrTxAdded := uint32(0)
	nrMiniBlocksProcessed := 0

	if check.IfNil(hdr) {
		return miniBlocks, nrTxAdded, false, nil
	}

	crossMiniBlockHashes := hdr.GetMiniBlockHeadersWithDst(tc.shardCoordinator.SelfId())
	for key, senderShardId := range crossMiniBlockHashes {
		if !haveTime() {
			log.Debug("CreateMbsAndProcessCrossShardTransactionsDstMe",
				"stop creating", "time is out")
			break
		}

		if tc.blockSizeComputation.IsMaxBlockSizeReached(0, 0) {
			log.Debug("CreateMbsAndProcessCrossShardTransactionsDstMe",
				"stop creating", "max block size has been reached")
			break
		}

		_, ok := processedMiniBlocksHashes[key]
		if ok {
			nrMiniBlocksProcessed++
			continue
		}

		miniVal, _ := tc.miniBlockPool.Peek([]byte(key))
		if miniVal == nil {
			go tc.onRequestMiniBlock(senderShardId, []byte(key))
			continue
		}

		miniBlock, ok := miniVal.(*block.MiniBlock)
		if !ok {
			continue
		}

		preproc := tc.getPreProcessor(miniBlock.Type)
		if check.IfNil(preproc) {
			return nil, 0, false, fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, miniBlock.Type)
		}

		requestedTxs := preproc.RequestTransactionsForMiniBlock(miniBlock)
		if requestedTxs > 0 {
			continue
		}

		err := tc.processCompleteMiniBlock(preproc, miniBlock, haveTime)
		if err != nil {
			continue
		}

		// all txs processed, add to processed miniblocks
		miniBlocks = append(miniBlocks, miniBlock)
		nrTxAdded = nrTxAdded + uint32(len(miniBlock.TxHashes))
		nrMiniBlocksProcessed++
	}

	allMBsProcessed := nrMiniBlocksProcessed == len(crossMiniBlockHashes)

	return miniBlocks, nrTxAdded, allMBsProcessed, nil
}

// CreateMbsAndProcessTransactionsFromMe creates miniblocks and processes transactions from pool
func (tc *transactionCoordinator) CreateMbsAndProcessTransactionsFromMe(
	haveTime func() bool,
) block.MiniBlockSlice {

	miniBlocks := make(block.MiniBlockSlice, 0)
	for _, blockType := range tc.keysTxPreProcs {
		txPreProc := tc.getPreProcessor(blockType)
		if check.IfNil(txPreProc) {
			return nil
		}

		mbs, err := txPreProc.CreateAndProcessMiniBlocks(haveTime)
		if err != nil {
			log.Debug("CreateAndProcessMiniBlocks", "error", err.Error())
		}

		if len(mbs) > 0 {
			miniBlocks = append(miniBlocks, mbs...)
		}
	}

	interMBs := tc.CreatePostProcessMiniBlocks()
	if len(interMBs) > 0 {
		miniBlocks = append(miniBlocks, interMBs...)
	}

	return miniBlocks
}

// CreatePostProcessMiniBlocks returns all the post processed miniblocks
func (tc *transactionCoordinator) CreatePostProcessMiniBlocks() block.MiniBlockSlice {
	miniBlocks := make(block.MiniBlockSlice, 0)

	// processing has to be done in order, as the order of different type of transactions over the same account is strict
	for _, blockType := range tc.keysInterimProcs {
		interimProc := tc.getInterimProcessor(blockType)
		if check.IfNil(interimProc) {
			continue
		}

		currMbs := interimProc.CreateAllInterMiniBlocks()
		for _, value := range currMbs {
			miniBlocks = append(miniBlocks, value)
		}
	}

	return miniBlocks
}

// CreateBlockStarted initializes necessary data for preprocessors at block create or block process
func (tc *transactionCoordinator) CreateBlockStarted() {
	tc.gasHandler.Init()
	tc.blockSizeComputation.Init()
	tc.balanceComputation.Init()

	tc.mutPreProcessor.RLock()
	for _, value := range tc.txPreProcessors {
		value.CreateBlockStarted()
	}
	tc.mutPreProcessor.RUnlock()

	tc.mutInterimProcessors.RLock()
	for _, value := range tc.interimProcessors {
		value.CreateBlockStarted()
	}
	tc.mutInterimProcessors.RUnlock()
}

func (tc *transactionCoordinator) getPreProcessor(blockType block.Type) process.PreProcessor {
	tc.mutPreProcessor.RLock()
	preprocessor, exists := tc.txPreProcessors[blockType]
	tc.mutPreProcessor.RUnlock()

	if !exists {
		return nil
	}

	return preprocessor
}

func (tc *transactionCoordinator) getInterimProcessor(blockType block.Type) process.IntermediateTransactionHandler {
	tc.mutInterimProcessors.RLock()
	interProcessor, exists := tc.interimProcessors[blockType]
	tc.mutInterimProcessors.RUnlock()

	if !exists {
		return nil
	}

	return interProcessor
}

func createBroadcastTopic(shardC sharding.Coordinator, destShId uint32, mbType block.Type) (string, error) {
	var baseTopic string

	switch mbType {
	case block.TxBlock:
		baseTopic = factory.TransactionTopic
	case block.PeerBlock:
		baseTopic = factory.PeerChBodyTopic
	case block.SmartContractResultBlock:
		baseTopic = factory.UnsignedTransactionTopic
	case block.RewardsBlock:
		baseTopic = factory.RewardsTransactionTopic
	default:
		return "", process.ErrUnknownBlockType
	}

	transactionTopic := baseTopic +
		shardC.CommunicationIdentifier(destShId)

	return transactionTopic, nil
}

// CreateMarshalizedData creates marshalized data for broadcasting
func (tc *transactionCoordinator) CreateMarshalizedData(body *block.Body) map[string][][]byte {
	mrsTxs := make(map[string][][]byte)

	if check.IfNil(body) {
		return mrsTxs
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock.SenderShardID != tc.shardCoordinator.SelfId() ||
			miniBlock.ReceiverShardID == tc.shardCoordinator.SelfId() {
			continue
		}

		broadcastTopic, err := createBroadcastTopic(tc.shardCoordinator, miniBlock.ReceiverShardID, miniBlock.Type)
		if err != nil {
			log.Warn("CreateMarshalizedData.createBroadcastTopic", "error", err.Error())
			continue
		}

		isPreProcessMiniBlock := miniBlock.Type == block.TxBlock
		preproc := tc.getPreProcessor(miniBlock.Type)
		if !check.IfNil(preproc) && isPreProcessMiniBlock {
			dataMarshalizer, ok := preproc.(process.DataMarshalizer)
			if ok {
				//preproc supports marshalizing items
				tc.appendMarshalizedItems(
					dataMarshalizer,
					miniBlock.TxHashes,
					mrsTxs,
					broadcastTopic,
				)
			}
		}

		interimProc := tc.getInterimProcessor(miniBlock.Type)
		if !check.IfNil(interimProc) && !isPreProcessMiniBlock {
			dataMarshalizer, ok := interimProc.(process.DataMarshalizer)
			if ok {
				//interimProc supports marshalizing items
				tc.appendMarshalizedItems(
					dataMarshalizer,
					miniBlock.TxHashes,
					mrsTxs,
					broadcastTopic,
				)
			}
		}
	}

	return mrsTxs
}

func (tc *transactionCoordinator) appendMarshalizedItems(
	dataMarshalizer process.DataMarshalizer,
	txHashes [][]byte,
	mrsTxs map[string][][]byte,
	broadcastTopic string,
) {
	currMrsTxs, err := dataMarshalizer.CreateMarshalizedData(txHashes)
	if err != nil {
		log.Debug("appendMarshalizedItems.CreateMarshalizedData", "error", err.Error())
		return
	}

	if len(currMrsTxs) > 0 {
		mrsTxs[broadcastTopic] = append(mrsTxs[broadcastTopic], currMrsTxs...)
	}
}

// GetAllCurrentUsedTxs returns the cached transaction data for current round
func (tc *transactionCoordinator) GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler {
	txPool := make(map[string]data.TransactionHandler)
	interTxPool := make(map[string]data.TransactionHandler)

	preProc := tc.getPreProcessor(blockType)
	if preProc != nil {
		txPool = preProc.GetAllCurrentUsedTxs()
	}

	interProc := tc.getInterimProcessor(blockType)
	if interProc != nil {
		interTxPool = interProc.GetAllCurrentFinishedTxs()
	}

	for hash, tx := range interTxPool {
		txPool[hash] = tx
	}

	return txPool
}

// RequestMiniBlocks request miniblocks if missing
func (tc *transactionCoordinator) RequestMiniBlocks(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	tc.requestedItemsHandler.Sweep()

	crossMiniBlockHashes := header.GetMiniBlockHeadersWithDst(tc.shardCoordinator.SelfId())
	for key, senderShardId := range crossMiniBlockHashes {
		obj, _ := tc.miniBlockPool.Peek([]byte(key))
		if obj == nil {
			go tc.onRequestMiniBlock(senderShardId, []byte(key))
			_ = tc.requestedItemsHandler.Add(key)
		}
	}
}

func (tc *transactionCoordinator) receivedMiniBlock(key []byte, value interface{}) {
	if key == nil {
		return
	}

	if !tc.requestedItemsHandler.Has(string(key)) {
		return
	}

	miniBlock, ok := value.(*block.MiniBlock)
	if !ok {
		log.Warn("transactionCoordinator.receivedMiniBlock", "error", process.ErrWrongTypeAssertion)
		return
	}

	log.Trace("transactionCoordinator.receivedMiniBlock", "hash", key)

	preproc := tc.getPreProcessor(miniBlock.Type)
	if check.IfNil(preproc) {
		log.Warn("transactionCoordinator.receivedMiniBlock",
			"error", fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, miniBlock.Type))
		return
	}

	numTxsRequested := preproc.RequestTransactionsForMiniBlock(miniBlock)
	if numTxsRequested > 0 {
		log.Debug("transactionCoordinator.receivedMiniBlock", "hash", key,
			"num txs requested", numTxsRequested)
	}
}

// processMiniBlockComplete - all transactions must be processed together, otherwise error
func (tc *transactionCoordinator) processCompleteMiniBlock(
	preproc process.PreProcessor,
	miniBlock *block.MiniBlock,
	haveTime func() bool,
) error {

	snapshot := tc.accounts.JournalLen()

	processedTxs, err := preproc.ProcessMiniBlock(miniBlock, haveTime)
	if err != nil {
		log.Debug("ProcessMiniBlock", "error", err.Error())

		errAccountState := tc.accounts.RevertToSnapshot(snapshot)
		if errAccountState != nil {
			// TODO: evaluate if reloading the trie from disk will might solve the problem
			log.Debug("RevertToSnapshot", "error", errAccountState.Error())
		}

		if len(processedTxs) > 0 {
			tc.revertProcessedTxsResults(processedTxs)
		}

		return err
	}

	return nil
}

func (tc *transactionCoordinator) revertProcessedTxsResults(txHashes [][]byte) {
	for _, value := range tc.keysInterimProcs {
		interProc, ok := tc.interimProcessors[value]
		if !ok {
			continue
		}
		interProc.RemoveProcessedResultsFor(txHashes)
	}
	tc.feeHandler.RevertFees(txHashes)
}

// VerifyCreatedBlockTransactions checks whether the created transactions are the same as the one proposed
func (tc *transactionCoordinator) VerifyCreatedBlockTransactions(hdr data.HeaderHandler, body *block.Body) error {
	tc.mutInterimProcessors.RLock()
	defer tc.mutInterimProcessors.RUnlock()
	errMutex := sync.Mutex{}
	var errFound error

	wg := sync.WaitGroup{}
	wg.Add(len(tc.interimProcessors))

	for _, interimProc := range tc.interimProcessors {
		go func(intermediateProcessor process.IntermediateTransactionHandler) {
			err := intermediateProcessor.VerifyInterMiniBlocks(body)
			if err != nil {
				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}(interimProc)
	}

	wg.Wait()

	if errFound != nil {
		return errFound
	}

	if check.IfNil(hdr) {
		return process.ErrNilBlockHeader
	}

	createdReceiptHash, err := tc.CreateReceiptsHash()
	if err != nil {
		return err
	}

	if !bytes.Equal(createdReceiptHash, hdr.GetReceiptsHash()) {
		return process.ErrReceiptsHashMissmatch
	}

	return nil
}

// CreateReceiptsHash will return the hash for the receipts
func (tc *transactionCoordinator) CreateReceiptsHash() ([]byte, error) {
	allReceiptsHashes := make([][]byte, 0)

	for _, value := range tc.keysInterimProcs {
		interProc, ok := tc.interimProcessors[value]
		if !ok {
			continue
		}

		mb := interProc.GetCreatedInShardMiniBlock()
		if mb == nil {
			log.Trace("CreateReceiptsHash nil inshard miniblock for type", "type", value)
			continue
		}

		log.Trace("CreateReceiptsHash.GetCreatedInShardMiniBlock",
			"type", mb.Type,
			"senderShardID", mb.SenderShardID,
			"receiverShardID", mb.ReceiverShardID,
			"numTxHashes", len(mb.TxHashes),
			"interimProcType", value,
		)

		for _, hash := range mb.TxHashes {
			log.Trace("tx", "hash", hash)
		}

		currHash, err := core.CalculateHash(tc.marshalizer, tc.hasher, mb)
		if err != nil {
			return nil, err
		}

		allReceiptsHashes = append(allReceiptsHashes, currHash)
	}

	finalReceiptHash, err := core.CalculateHash(tc.marshalizer, tc.hasher, &batch.Batch{Data: allReceiptsHashes})
	return finalReceiptHash, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *transactionCoordinator) IsInterfaceNil() bool {
	return tc == nil
}
