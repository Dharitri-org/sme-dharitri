package sync

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/process"
)

func (boot *ShardBootstrap) RequestHeaderWithNonce(nonce uint64) {
	boot.requestHeaderWithNonce(nonce)
}

func (boot *ShardBootstrap) GetMiniBlocks(hashes [][]byte) ([]*process.MiniblockAndHash, [][]byte) {
	return boot.miniBlocksProvider.GetMiniBlocks(hashes)
}

func (boot *MetaBootstrap) ReceivedHeaders(header data.HeaderHandler, key []byte) {
	boot.processReceivedHeader(header, key)
}

func (boot *ShardBootstrap) ReceivedHeaders(header data.HeaderHandler, key []byte) {
	boot.processReceivedHeader(header, key)
}

func (boot *ShardBootstrap) RollBack(revertUsingForkNonce bool) error {
	return boot.rollBack(revertUsingForkNonce)
}

func (boot *MetaBootstrap) RollBack(revertUsingForkNonce bool) error {
	return boot.rollBack(revertUsingForkNonce)
}

func (bfd *baseForkDetector) GetHeaders(nonce uint64) []*headerInfo {
	bfd.mutHeaders.Lock()
	defer bfd.mutHeaders.Unlock()

	headers := bfd.headers[nonce]

	if headers == nil {
		return nil
	}

	newHeaders := make([]*headerInfo, len(headers))
	copy(newHeaders, headers)

	return newHeaders
}

func (bfd *baseForkDetector) LastCheckpointNonce() uint64 {
	return bfd.lastCheckpoint().nonce
}

func (bfd *baseForkDetector) LastCheckpointRound() uint64 {
	return bfd.lastCheckpoint().round
}

func (bfd *baseForkDetector) SetFinalCheckpoint(nonce uint64, round uint64, hash []byte) {
	bfd.setFinalCheckpoint(&checkpointInfo{nonce: nonce, round: round, hash: hash})
}

func (bfd *baseForkDetector) FinalCheckpointNonce() uint64 {
	return bfd.finalCheckpoint().nonce
}

func (bfd *baseForkDetector) FinalCheckpointRound() uint64 {
	return bfd.finalCheckpoint().round
}

func (bfd *baseForkDetector) CheckBlockValidity(header *block.Header, headerHash []byte) error {
	return bfd.checkBlockBasicValidity(header, headerHash)
}

func (bfd *baseForkDetector) RemovePastHeaders() {
	bfd.removePastHeaders()
}

func (bfd *baseForkDetector) RemoveInvalidReceivedHeaders() {
	bfd.removeInvalidReceivedHeaders()
}

func (bfd *baseForkDetector) ComputeProbableHighestNonce() uint64 {
	return bfd.computeProbableHighestNonce()
}

func (bfd *baseForkDetector) IsConsensusStuck() bool {
	return bfd.isConsensusStuck()
}

func (hi *headerInfo) Hash() []byte {
	return hi.hash
}

func (hi *headerInfo) GetBlockHeaderState() process.BlockHeaderState {
	return hi.state
}

func (boot *ShardBootstrap) NotifySyncStateListeners() {
	isNodeSynchronized := boot.GetNodeState() == core.NsSynchronized
	boot.notifySyncStateListeners(isNodeSynchronized)
}

func (boot *MetaBootstrap) NotifySyncStateListeners() {
	isNodeSynchronized := boot.GetNodeState() == core.NsSynchronized
	boot.notifySyncStateListeners(isNodeSynchronized)
}

func (boot *ShardBootstrap) SyncStateListeners() []func(bool) {
	return boot.syncStateListeners
}

func (boot *MetaBootstrap) SyncStateListeners() []func(bool) {
	return boot.syncStateListeners
}

func (boot *ShardBootstrap) SetForkNonce(nonce uint64) {
	boot.forkInfo.Nonce = nonce
}

func (boot *MetaBootstrap) SetForkNonce(nonce uint64) {
	boot.forkInfo.Nonce = nonce
}

func (boot *ShardBootstrap) IsForkDetected() bool {
	return boot.forkInfo.IsDetected
}

func (boot *MetaBootstrap) IsForkDetected() bool {
	return boot.forkInfo.IsDetected
}

func (boot *MetaBootstrap) GetNotarizedInfo(
	lastNotarized map[uint32]*HdrInfo,
	finalNotarized map[uint32]*HdrInfo,
	blockWithLastNotarized map[uint32]uint64,
	blockWithFinalNotarized map[uint32]uint64,
	startNonce uint64,
) *notarizedInfo {
	return &notarizedInfo{
		lastNotarized:           lastNotarized,
		finalNotarized:          finalNotarized,
		blockWithLastNotarized:  blockWithLastNotarized,
		blockWithFinalNotarized: blockWithFinalNotarized,
		startNonce:              startNonce,
	}
}

func (boot *baseBootstrap) ProcessReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	boot.processReceivedHeader(headerHandler, headerHash)
}

func (boot *ShardBootstrap) RequestMiniBlocksFromHeaderWithNonceIfMissing(headerHandler data.HeaderHandler) {
	boot.requestMiniBlocksFromHeaderWithNonceIfMissing(headerHandler)
}

func (bfd *baseForkDetector) IsHeaderReceivedTooLate(header data.HeaderHandler, state process.BlockHeaderState, finality int64) bool {
	return bfd.isHeaderReceivedTooLate(header, state, finality)
}

func (bfd *baseForkDetector) SetProbableHighestNonce(nonce uint64) {
	bfd.setProbableHighestNonce(nonce)
}

func (sfd *shardForkDetector) ComputeFinalCheckpoint() {
	sfd.computeFinalCheckpoint()
}

func (bfd *baseForkDetector) AddCheckPoint(round uint64, nonce uint64, hash []byte) {
	bfd.addCheckpoint(&checkpointInfo{round: round, nonce: nonce, hash: hash})
}

func (bfd *baseForkDetector) ComputeGenesisTimeFromHeader(headerHandler data.HeaderHandler) int64 {
	return bfd.computeGenesisTimeFromHeader(headerHandler)
}

func (boot *baseBootstrap) InitNotarizedMap() map[uint32]*HdrInfo {
	return make(map[uint32]*HdrInfo)
}

func (boot *baseBootstrap) SetNotarizedMap(notarizedMap map[uint32]*HdrInfo, shardId uint32, nonce uint64, hash []byte) {
	hdrInfo, ok := notarizedMap[shardId]
	if !ok {
		notarizedMap[shardId] = &HdrInfo{Nonce: nonce, Hash: hash}
		return
	}

	hdrInfo.Nonce = nonce
	hdrInfo.Hash = hash
}

func (boot *baseBootstrap) SetNodeStateCalculated(state bool) {
	boot.mutNodeState.Lock()
	boot.isNodeStateCalculated = state
	boot.mutNodeState.Unlock()
}

func (boot *baseBootstrap) ComputeNodeState() {
	boot.computeNodeState()
}

func (boot *baseBootstrap) DoJobOnSyncBlockFail(bodyHandler data.BodyHandler, headerHandler data.HeaderHandler, err error) {
	boot.doJobOnSyncBlockFail(bodyHandler, headerHandler, err)
}

func (boot *baseBootstrap) SetNumSyncedWithErrorsForNonce(nonce uint64, numSyncedWithErrors uint32) {
	boot.mutNonceSyncedWithErrors.Lock()
	boot.mapNonceSyncedWithErrors[nonce] = numSyncedWithErrors
	boot.mutNonceSyncedWithErrors.Unlock()
}

func (boot *baseBootstrap) GetNumSyncedWithErrorsForNonce(nonce uint64) uint32 {
	boot.mutNonceSyncedWithErrors.RLock()
	numSyncedWithErrors := boot.mapNonceSyncedWithErrors[nonce]
	boot.mutNonceSyncedWithErrors.RUnlock()

	return numSyncedWithErrors
}

func (boot *baseBootstrap) GetMapNonceSyncedWithErrorsLen() int {
	boot.mutNonceSyncedWithErrors.RLock()
	mapNonceSyncedWithErrorsLen := len(boot.mapNonceSyncedWithErrors)
	boot.mutNonceSyncedWithErrors.RUnlock()

	return mapNonceSyncedWithErrorsLen
}

func (boot *baseBootstrap) CleanNoncesSyncedWithErrorsBehindFinal() {
	boot.cleanNoncesSyncedWithErrorsBehindFinal()
}
