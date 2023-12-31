package indexer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/core/statistics"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/rewardTx"
	"github.com/Dharitri-org/sme-dharitri/data/smartContractResult"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/marshal"
)

type commonProcessor struct {
	addressPubkeyConverter   core.PubkeyConverter
	validatorPubkeyConverter core.PubkeyConverter
}

func checkElasticSearchParams(arguments ElasticIndexerArgs) error {
	if check.IfNil(arguments.AddressPubkeyConverter) {
		return fmt.Errorf("%w when setting addressPubkeyConverter in indexer", ErrNilPubkeyConverter)
	}
	if check.IfNil(arguments.ValidatorPubkeyConverter) {
		return fmt.Errorf("%w when setting validatorPubkeyConverter in indexer", ErrNilPubkeyConverter)
	}
	if arguments.Url == "" {
		return core.ErrNilUrl
	}
	if arguments.UserName == "" {
		return ErrEmptyUserName
	}
	if arguments.Password == "" {
		return ErrEmptyPassword
	}
	if check.IfNil(arguments.Marshalizer) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return core.ErrNilHasher
	}
	if check.IfNil(arguments.NodesCoordinator) {
		return core.ErrNilNodesCoordinator
	}
	if arguments.EpochStartNotifier == nil {
		return core.ErrNilEpochStartNotifier
	}

	return nil
}

func timestampMapping() io.Reader {
	return strings.NewReader(
		`{
				"settings": {"index": {"sort.field": "timestamp", "sort.order": "desc"}},
				"mappings": {"properties": {"timestamp": {"type": "date"}}}
			}`,
	)
}

func prepareGeneralInfo(tpsBenchmark statistics.TPSBenchmark) bytes.Buffer {
	var buff bytes.Buffer

	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, metachainTpsDocID, tpsIndex, "\n"))
	generalInfo := TPS{
		LiveTPS:               tpsBenchmark.LiveTPS(),
		PeakTPS:               tpsBenchmark.PeakTPS(),
		NrOfShards:            tpsBenchmark.NrOfShards(),
		BlockNumber:           tpsBenchmark.BlockNumber(),
		RoundNumber:           tpsBenchmark.RoundNumber(),
		RoundTime:             tpsBenchmark.RoundTime(),
		AverageBlockTxCount:   tpsBenchmark.AverageBlockTxCount(),
		LastBlockTxCount:      tpsBenchmark.LastBlockTxCount(),
		TotalProcessedTxCount: tpsBenchmark.TotalProcessedTxCount(),
	}

	serializedInfo, err := json.Marshal(generalInfo)
	if err != nil {
		log.Debug("indexer: could not serialize tps info, will skip indexing tps this round")
		return buff
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	buff.Grow(len(meta) + len(serializedInfo))
	_, err = buff.Write(meta)
	if err != nil {
		log.Warn("elastic search: update TPS write meta", "error", err.Error())
	}
	_, err = buff.Write(serializedInfo)
	if err != nil {
		log.Warn("elastic search: update TPS write serialized info", "error", err.Error())
	}

	return buff
}

func serializeShardInfo(shardInfo statistics.ShardStatistic) ([]byte, []byte) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s%d", "_type" : "%s" } }%s`,
		shardTpsDocIDPrefix, shardInfo.ShardID(), tpsIndex, "\n"))

	bigTxCount := big.NewInt(int64(shardInfo.AverageBlockTxCount()))
	shardTPS := TPS{
		ShardID:               shardInfo.ShardID(),
		LiveTPS:               shardInfo.LiveTPS(),
		PeakTPS:               shardInfo.PeakTPS(),
		AverageTPS:            shardInfo.AverageTPS(),
		AverageBlockTxCount:   bigTxCount,
		CurrentBlockNonce:     shardInfo.CurrentBlockNonce(),
		LastBlockTxCount:      shardInfo.LastBlockTxCount(),
		TotalProcessedTxCount: shardInfo.TotalProcessedTxCount(),
	}

	serializedInfo, err := json.Marshal(shardTPS)
	if err != nil {
		log.Debug("indexer: could not serialize tps info, will skip indexing tps this shard")
		return nil, nil
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
}

func (cm *commonProcessor) buildTransaction(
	tx *transaction.Transaction,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
	txStatus string,
) *Transaction {
	return &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		Nonce:         tx.Nonce,
		Round:         header.GetRound(),
		Value:         tx.Value.String(),
		Receiver:      cm.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:        cm.addressPubkeyConverter.Encode(tx.SndAddr),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      tx.GasPrice,
		GasLimit:      tx.GasLimit,
		Data:          tx.Data,
		Signature:     hex.EncodeToString(tx.Signature),
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        txStatus,
		GasUsed:       tx.GasLimit,
	}
}

func (cm *commonProcessor) buildRewardTransaction(
	rTx *rewardTx.RewardTx,
	txHash []byte,
	mbHash []byte,
	mb *block.MiniBlock,
	header data.HeaderHandler,
	txStatus string,
) *Transaction {
	return &Transaction{
		Hash:          hex.EncodeToString(txHash),
		MBHash:        hex.EncodeToString(mbHash),
		Nonce:         0,
		Round:         rTx.Round,
		Value:         rTx.Value.String(),
		Receiver:      cm.addressPubkeyConverter.Encode(rTx.RcvAddr),
		Sender:        fmt.Sprintf("%d", core.MetachainShardId),
		ReceiverShard: mb.ReceiverShardID,
		SenderShard:   mb.SenderShardID,
		GasPrice:      0,
		GasLimit:      0,
		Data:          make([]byte, 0),
		Signature:     "",
		Timestamp:     time.Duration(header.GetTimeStamp()),
		Status:        txStatus,
	}
}

func (cm *commonProcessor) convertScResultInDatabaseScr(sc *smartContractResult.SmartContractResult) ScResult {
	decodedData := decodeScResultData(sc.Data)
	return ScResult{
		Nonce:         sc.Nonce,
		GasLimit:      sc.GasLimit,
		GasPrice:      sc.GasPrice,
		Value:         sc.Value.String(),
		Sender:        cm.addressPubkeyConverter.Encode(sc.SndAddr),
		Receiver:      cm.addressPubkeyConverter.Encode(sc.RcvAddr),
		Code:          string(sc.Code),
		Data:          decodedData,
		PreTxHash:     hex.EncodeToString(sc.PrevTxHash),
		CallType:      strconv.Itoa(int(sc.CallType)),
		CodeMetadata:  sc.CodeMetadata,
		ReturnMessage: string(sc.ReturnMessage),
	}
}

func decodeScResultData(scrData []byte) []byte {
	encodedData := strings.Split(string(scrData), "@")
	encodedData = append([]string(nil), encodedData[1:]...)

	decodedData := ""
	for i, enc := range encodedData {
		if !canInterpretAsString([]byte(enc)) {
			continue
		}
		if i > 0 {
			decodedData += "@" + enc
			continue
		}

		val, _ := hex.DecodeString(enc)
		if len(val) > 0 {
			decodedData += "@" + string(val)
		}
	}

	return []byte(decodedData)
}

func canInterpretAsString(bytes []byte) bool {
	if len(bytes) == 0 {
		return false
	}
	for _, b := range bytes {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}

func serializeBulkMiniBlocks(
	hdrShardID uint32,
	bulkMbs []*Miniblock,
	getAlreadyIndexedItems func(hashes []string, index string) (map[string]bool, error),
) (bytes.Buffer, map[string]bool) {
	var err error
	var buff bytes.Buffer

	mbsHashes := make([]string, len(bulkMbs))
	for idx := range bulkMbs {
		mbsHashes[idx] = bulkMbs[idx].Hash
	}

	existsInDb, err := getAlreadyIndexedItems(mbsHashes, miniblocksIndex)
	if err != nil {
		log.Warn("indexer get indexed items miniblocks",
			"error", err.Error())
		return buff, make(map[string]bool)
	}

	for _, mb := range bulkMbs {
		var meta, serializedData []byte
		if !existsInDb[mb.Hash] {
			//insert miniblock in database
			meta = []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, mb.Hash, "_doc", "\n"))
			serializedData, err = json.Marshal(mb)
			if err != nil {
				log.Debug("indexer: marshal",
					"error", "could not serialize miniblock, will skip indexing",
					"mb hash", mb.Hash)
				continue
			}
		} else {
			// update miniblock
			meta = []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s" } }%s`, mb.Hash, "\n"))
			if hdrShardID == mb.SenderShardID {
				// update sender block hash
				serializedData = []byte(fmt.Sprintf(`{ "doc" : { "senderBlockHash" : "%s" } }`, mb.SenderBlockHash))
			} else {
				// update receiver block hash
				serializedData = []byte(fmt.Sprintf(`{ "doc" : { "receiverBlockHash" : "%s" } }`, mb.ReceiverBlockHash))
			}
		}

		buff = prepareBufferMiniblocks(buff, meta, serializedData)
	}

	return buff, existsInDb
}

func prepareBufferMiniblocks(buff bytes.Buffer, meta, serializedData []byte) bytes.Buffer {
	// append a newline for each element
	serializedData = append(serializedData, "\n"...)
	buff.Grow(len(meta) + len(serializedData))
	_, err := buff.Write(meta)
	if err != nil {
		log.Warn("elastic search: serialize bulk miniblocks, write meta", "error", err.Error())
	}
	_, err = buff.Write(serializedData)
	if err != nil {
		log.Warn("elastic search: serialize bulk miniblocks, write serialized miniblock", "error", err.Error())
	}

	return buff
}

func serializeTransactions(
	transactions []*Transaction,
	selfShardID uint32,
	_ func(hashes []string, index string) (map[string]bool, error),
	mbsHashInDB map[string]bool,
) []bytes.Buffer {
	var err error

	var buff bytes.Buffer
	buffSlice := make([]bytes.Buffer, 0)
	for _, tx := range transactions {
		isMBOfTxInDB := mbsHashInDB[tx.MBHash]
		meta, serializedData := prepareSerializedDataForATransaction(tx, selfShardID, isMBOfTxInDB)
		if len(meta) == 0 {
			continue
		}

		// append a newline for each element
		serializedData = append(serializedData, "\n"...)

		buffLenWithCurrentTx := buff.Len() + len(meta) + len(serializedData)
		if buffLenWithCurrentTx > txsBulkSizeThreshold && buff.Len() != 0 {
			buffSlice = append(buffSlice, buff)
			buff = bytes.Buffer{}
		}

		buff.Grow(len(meta) + len(serializedData))
		_, err = buff.Write(meta)
		if err != nil {
			log.Warn("elastic search: serialize bulk tx, write meta", "error", err.Error())
		}
		_, err = buff.Write(serializedData)
		if err != nil {
			log.Warn("elastic search: serialize bulk tx, write serialized tx", "error", err.Error())
		}
	}

	// check if the last buffer contains data
	if buff.Len() != 0 {
		buffSlice = append(buffSlice, buff)
	}

	return buffSlice
}

func prepareSerializedDataForATransaction(
	tx *Transaction,
	selfShardID uint32,
	isMBOfTxInDB bool,
) (meta []byte, serializedData []byte) {
	var err error
	if isMBOfTxInDB {
		if !isCrossShardDstMe(tx, selfShardID) || tx.Status == txStatusInvalid {
			return
		}
		// update transaction
		meta, serializedData = prepareTxUpdate(tx)
	} else {
		// insert transaction in database
		meta = []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, tx.Hash, "\n"))
		serializedData, err = json.Marshal(tx)
		if err != nil {
			log.Debug("indexer: marshal",
				"error", "could not serialize transaction, will skip indexing",
				"tx hash", tx.Hash)
			return
		}
	}
	return
}

func prepareTxUpdate(tx *Transaction) ([]byte, []byte) {
	var meta, serializedData []byte

	meta = []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s", "_type" : "%s"  } }%s`, tx.Hash, "_doc", "\n"))

	marshalizedLog, err := json.Marshal(tx.Log)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize transaction log, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil
	}
	scResults, err := json.Marshal(tx.SmartContractResults)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize smart contract results, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil
	}

	marshalizedTimestamp, err := json.Marshal(tx.Timestamp)
	if err != nil {
		log.Debug("indexer: marshal",
			"error", "could not serialize timestamp, will skip indexing",
			"tx hash", tx.Hash)
		return nil, nil
	}

	if tx.GasUsed == tx.GasLimit {
		// do not update gasUsed because it is the same with gasUsed when transaction was saved first time in database
		serializedData = []byte(fmt.Sprintf(`{ "doc" : { "log" : %s, "scResults" : %s, "status": "%s", "timestamp": %s } }`,
			string(marshalizedLog), string(scResults), tx.Status, string(marshalizedTimestamp)))
	} else {
		// update gasUsed because was changed (is a smart contract operation)
		serializedData = []byte(fmt.Sprintf(`{ "doc" : { "log" : %s, "scResults" : %s, "status": "%s", "timestamp": %s, "gasUsed" : %s } }`,
			string(marshalizedLog), string(scResults), tx.Status, string(marshalizedTimestamp), fmt.Sprintf("%d", tx.GasUsed)))
	}

	return meta, serializedData
}

func isCrossShardDstMe(tx *Transaction, selfShardID uint32) bool {
	return tx.SenderShard != tx.ReceiverShard && tx.ReceiverShard == selfShardID
}

func computeSizeOfTxs(marshalizer marshal.Marshalizer, txs map[string]data.TransactionHandler) int {
	if len(txs) == 0 {
		return 0
	}

	txsSize := 0
	for _, tx := range txs {
		txBytes, err := marshalizer.Marshal(tx)
		if err != nil {
			log.Debug("indexer: marshal transaction", "error", err)
			continue
		}

		txsSize += len(txBytes)
	}

	return txsSize
}

func getDecodedResponseMultiGet(response object) map[string]bool {
	founded := make(map[string]bool)
	interfaceSlice, ok := response["docs"].([]interface{})
	if !ok {
		return founded
	}

	for _, element := range interfaceSlice {
		obj := element.(object)
		_, ok = obj["error"]
		if ok {
			continue
		}
		founded[obj["_id"].(string)] = obj["found"].(bool)
	}

	return founded
}
