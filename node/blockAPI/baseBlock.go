package blockAPI

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core/fullHistory"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/data/typeConverters"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	logger "github.com/Dharitri-org/sme-logger"
)

type baseAPIBockProcessor struct {
	isFullHistoryNode        bool
	selfShardID              uint32
	store                    dataRetriever.StorageService
	marshalizer              marshal.Marshalizer
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	historyRepo              fullHistory.HistoryRepository
	unmarshalTx              func(txBytes []byte, txType string) (*transaction.ApiTransactionResult, error)
}

var log = logger.GetOrCreate("node/blockAPI")

func (bap *baseAPIBockProcessor) getTxsByMb(mbHeader *block.MiniBlockHeader, epoch uint32) []*transaction.ApiTransactionResult {
	mbBytes, err := bap.getFromStorerWithEpoch(dataRetriever.MiniBlockUnit, mbHeader.Hash, epoch)
	if err != nil {
		log.Warn("cannot get miniblock from storage",
			"hash", hex.EncodeToString(mbHeader.Hash),
			"error", err.Error())
		return nil
	}

	miniBlock := &block.MiniBlock{}
	err = bap.marshalizer.Unmarshal(miniBlock, mbBytes)
	if err != nil {
		log.Warn("cannot unmarshal miniblock",
			"hash", hex.EncodeToString(mbHeader.Hash),
			"error", err.Error())
		return nil
	}

	switch miniBlock.Type {
	case block.TxBlock:
		return bap.getTxsFromMiniblock(miniBlock, epoch, "normal", dataRetriever.TransactionUnit)
	case block.RewardsBlock:
		return bap.getTxsFromMiniblock(miniBlock, epoch, "reward", dataRetriever.RewardTransactionUnit)
	case block.SmartContractResultBlock:
		return bap.getTxsFromMiniblock(miniBlock, epoch, "unsigned", dataRetriever.UnsignedTransactionUnit)
	default:
		return nil
	}
}

func (bap *baseAPIBockProcessor) getTxsFromMiniblock(
	miniblock *block.MiniBlock,
	epoch uint32,
	txType string,
	unit dataRetriever.UnitType,
) []*transaction.ApiTransactionResult {
	storer := bap.store.GetStorer(unit)
	start := time.Now()
	marshalizedTxs, err := storer.GetBulkFromEpoch(miniblock.TxHashes, epoch)
	if err != nil {
		log.Warn("cannot get from storage transactions",
			"error", err.Error())
		return []*transaction.ApiTransactionResult{}
	}
	log.Debug(fmt.Sprintf("GetBulkFromEpoch took %s", time.Since(start)))

	start = time.Now()
	txs := make([]*transaction.ApiTransactionResult, 0)
	for txHash, txBytes := range marshalizedTxs {
		tx, err := bap.unmarshalTx(txBytes, txType)
		if err != nil {
			log.Warn("cannot unmarshal transaction",
				"hash", hex.EncodeToString([]byte(txHash)),
				"error", err.Error())
			continue
		}
		tx.Hash = hex.EncodeToString([]byte(txHash))

		txs = append(txs, tx)
	}
	log.Debug(fmt.Sprintf("UnmarshalTransactions took %s", time.Since(start)))

	return txs
}

func (bap *baseAPIBockProcessor) getFromStorer(unit dataRetriever.UnitType, key []byte) ([]byte, error) {
	if !bap.isFullHistoryNode {
		return bap.store.Get(unit, key)
	}

	epoch, err := bap.historyRepo.GetEpochForHash(key)
	if err != nil {
		return nil, err
	}

	storer := bap.store.GetStorer(unit)
	return storer.GetFromEpoch(key, epoch)
}

func (bap *baseAPIBockProcessor) getFromStorerWithEpoch(unit dataRetriever.UnitType, key []byte, epoch uint32) ([]byte, error) {
	storer := bap.store.GetStorer(unit)
	return storer.GetFromEpoch(key, epoch)
}
