package indexer

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/indexer/disabled"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/receipt"
	"github.com/Dharitri-org/sme-dharitri/data/rewardTx"
	"github.com/Dharitri-org/sme-dharitri/data/smartContractResult"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
)

const (
	txStatusSuccess     = "Success"
	txStatusPending     = "Pending"
	txStatusInvalid     = "Invalid"
	txStatusNotExecuted = "Not Executed"
	// A smart contract action (deploy, call, ...) should have minimum 2 smart contract results
	// exception to this rule are smart contract calls to DCT contract
	minimumNumberOfSmartContractResults = 2
)

type txDatabaseProcessor struct {
	*commonProcessor
	txLogsProcessor process.TransactionLogProcessorDatabase
	hasher          hashing.Hasher
	marshalizer     marshal.Marshalizer
}

func newTxDatabaseProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter core.PubkeyConverter,
	validatorPubkeyConverter core.PubkeyConverter,
) *txDatabaseProcessor {
	return &txDatabaseProcessor{
		hasher:      hasher,
		marshalizer: marshalizer,
		commonProcessor: &commonProcessor{
			addressPubkeyConverter:   addressPubkeyConverter,
			validatorPubkeyConverter: validatorPubkeyConverter,
		},
		txLogsProcessor: disabled.NewNilTxLogsProcessor(),
	}
}

func (tdp *txDatabaseProcessor) prepareTransactionsForDatabase(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	selfShardID uint32,
) []*Transaction {
	transactions, rewardsTxs := tdp.groupNormalTxsAndRewards(body, txPool, header, selfShardID)
	receipts := groupReceipts(txPool)
	scResults := groupSmartContractResults(txPool)

	for _, rec := range receipts {
		tx, ok := transactions[string(rec.TxHash)]
		if !ok {
			continue
		}

		gasUsed := big.NewInt(0).SetUint64(tx.GasPrice)
		gasUsed.Mul(gasUsed, big.NewInt(0).SetUint64(tx.GasLimit))
		gasUsed.Sub(gasUsed, rec.Value)
		gasUsed.Div(gasUsed, big.NewInt(0).SetUint64(tx.GasPrice))

		tx.GasUsed = gasUsed.Uint64()
	}

	countScResults := make(map[string]int)
	for _, scResult := range scResults {
		tx, ok := transactions[string(scResult.OriginalTxHash)]
		if !ok {
			continue
		}

		tx = tdp.addScResultInfoInTx(scResult, tx)

		countScResults[string(scResult.OriginalTxHash)]++
	}

	for hash, nrScResult := range countScResults {
		if nrScResult < minimumNumberOfSmartContractResults {
			if len(transactions[hash].SmartContractResults) > 0 {
				scResultData := transactions[hash].SmartContractResults[0].Data
				if bytes.Contains(scResultData, []byte("@ok")) {
					// DCT contract calls generate just one smart contract result
					continue
				}
			}

			transactions[hash].Status = txStatusNotExecuted
		}
	}

	// TODO for the moment do not save logs in database
	// uncomment this when transaction logs need to be saved in database
	//for hash, tx := range transactions {
	//	txLog, ok := tdp.txLogsProcessor.GetLogFromCache([]byte(hash))
	//	if !ok {
	//		continue
	//	}
	//
	//	tx.Log = tdp.prepareTxLog(txLog)
	//}

	tdp.txLogsProcessor.Clean()

	return append(convertMapTxsToSlice(transactions), rewardsTxs...)
}

func (tdp *txDatabaseProcessor) addScResultInfoInTx(scr *smartContractResult.SmartContractResult, tx *Transaction) *Transaction {
	dbScResult := tdp.commonProcessor.convertScResultInDatabaseScr(scr)
	tx.SmartContractResults = append(tx.SmartContractResults, dbScResult)

	if isSCRForSenderWithGasUsed(dbScResult, tx) {
		gasUsed := tx.GasLimit - scr.GasLimit
		tx.GasUsed = gasUsed
	}

	return tx
}

func isSCRForSenderWithGasUsed(dbScResult ScResult, tx *Transaction) bool {
	isForSender := dbScResult.Receiver == tx.Sender
	isWithGasLimit := dbScResult.GasLimit != 0
	isFromCurrentTx := dbScResult.PreTxHash == tx.Hash

	return isFromCurrentTx && isForSender && isWithGasLimit
}

func (tdp *txDatabaseProcessor) prepareTxLog(log data.LogHandler) TxLog {
	scAddr := tdp.addressPubkeyConverter.Encode(log.GetAddress())
	events := log.GetLogEvents()

	txLogEvents := make([]Event, len(events))
	for i, event := range events {
		txLogEvents[i].Address = hex.EncodeToString(event.GetAddress())
		txLogEvents[i].Data = hex.EncodeToString(event.GetData())
		txLogEvents[i].Identifier = hex.EncodeToString(event.GetIdentifier())

		topics := event.GetTopics()
		txLogEvents[i].Topics = make([]string, len(topics))
		for j, topic := range topics {
			txLogEvents[i].Topics[j] = hex.EncodeToString(topic)
		}
	}

	return TxLog{
		Address: scAddr,
		Events:  txLogEvents,
	}
}

func convertMapTxsToSlice(txs map[string]*Transaction) []*Transaction {
	transactions := make([]*Transaction, len(txs))
	i := 0
	for _, tx := range txs {
		transactions[i] = tx
		i++
	}
	return transactions
}

func (tdp *txDatabaseProcessor) groupNormalTxsAndRewards(
	body *block.Body,
	txPool map[string]data.TransactionHandler,
	header data.HeaderHandler,
	selfShardID uint32,
) (
	map[string]*Transaction,
	[]*Transaction,
) {
	transactions := make(map[string]*Transaction)
	rewardsTxs := make([]*Transaction, 0)

	for _, mb := range body.MiniBlocks {
		mbHash, err := core.CalculateHash(tdp.marshalizer, tdp.hasher, mb)
		if err != nil {
			continue
		}

		mbTxStatus := txStatusPending
		if selfShardID == mb.ReceiverShardID {
			mbTxStatus = txStatusSuccess
		}

		switch mb.Type {
		case block.TxBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, mb, header, mbTxStatus)
				transactions[hash] = dbTx
				delete(txPool, hash)
			}
		case block.InvalidBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, mb, header, txStatusInvalid)
				transactions[hash] = dbTx
				delete(txPool, hash)
			}
		case block.RewardsBlock:
			rTxs := getRewardsTransaction(txPool, mb.TxHashes)
			for hash, rtx := range rTxs {
				dbTx := tdp.commonProcessor.buildRewardTransaction(rtx, []byte(hash), mbHash, mb, header, mbTxStatus)
				rewardsTxs = append(rewardsTxs, dbTx)
				delete(txPool, hash)
			}
		default:
			continue
		}
	}

	return transactions, rewardsTxs
}

func groupSmartContractResults(txPool map[string]data.TransactionHandler) []*smartContractResult.SmartContractResult {
	scResults := make([]*smartContractResult.SmartContractResult, 0)
	for _, tx := range txPool {
		scResult, ok := tx.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		scResults = append(scResults, scResult)
	}

	return scResults
}

func groupReceipts(txPool map[string]data.TransactionHandler) []*receipt.Receipt {
	receipts := make([]*receipt.Receipt, 0)
	for hash, tx := range txPool {
		rec, ok := tx.(*receipt.Receipt)
		if !ok {
			continue
		}

		receipts = append(receipts, rec)
		delete(txPool, hash)
	}

	return receipts
}

func getTransactions(txPool map[string]data.TransactionHandler,
	txHashes [][]byte,
) map[string]*transaction.Transaction {
	transactions := make(map[string]*transaction.Transaction)
	for _, txHash := range txHashes {
		txHandler, ok := txPool[string(txHash)]
		if !ok {
			continue
		}

		tx, ok := txHandler.(*transaction.Transaction)
		if !ok {
			continue
		}
		transactions[string(txHash)] = tx
	}
	return transactions
}

func getRewardsTransaction(txPool map[string]data.TransactionHandler,
	txHashes [][]byte,
) map[string]*rewardTx.RewardTx {
	rewardsTxs := make(map[string]*rewardTx.RewardTx)
	for _, txHash := range txHashes {
		txHandler, ok := txPool[string(txHash)]
		if !ok {
			continue
		}

		reward, ok := txHandler.(*rewardTx.RewardTx)
		if !ok {
			continue
		}
		rewardsTxs[string(txHash)] = reward
	}
	return rewardsTxs
}
