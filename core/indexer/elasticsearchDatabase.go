package indexer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/statistics"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// elasticSearchDatabaseArgs is struct that is used to store all parameters that are needed to create a elasticsearch database
type elasticSearchDatabaseArgs struct {
	url                      string
	userName                 string
	password                 string
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	addressPubkeyConverter   core.PubkeyConverter
	validatorPubkeyConverter core.PubkeyConverter
}

// elasticSearchDatabase object it contains business logic built over databaseWriterHandler glue code wrapper
type elasticSearchDatabase struct {
	*txDatabaseProcessor
	dbClient    databaseClientHandler
	marshalizer marshal.Marshalizer
	hasher      hashing.Hasher
}

// newElasticSearchDatabase is method that will create a new elastic search dbClient
func newElasticSearchDatabase(arguments elasticSearchDatabaseArgs) (*elasticSearchDatabase, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{arguments.url},
		Username:  arguments.userName,
		Password:  arguments.password,
	}
	es, err := newDatabaseWriter(cfg)
	if err != nil {
		return nil, err
	}

	esdb := &elasticSearchDatabase{
		dbClient:    es,
		marshalizer: arguments.marshalizer,
		hasher:      arguments.hasher,
	}
	esdb.txDatabaseProcessor = newTxDatabaseProcessor(
		arguments.hasher,
		arguments.marshalizer,
		arguments.addressPubkeyConverter,
		arguments.validatorPubkeyConverter,
	)

	err = esdb.createIndexes()
	if err != nil {
		return nil, err
	}

	return esdb, nil
}

func (esd *elasticSearchDatabase) createIndexes() error {
	err := esd.dbClient.CheckAndCreateIndex(blockIndex, timestampMapping())
	if err != nil {
		return err
	}

	err = esd.dbClient.CheckAndCreateIndex(txIndex, timestampMapping())
	if err != nil {
		return err
	}

	err = esd.dbClient.CheckAndCreateIndex(tpsIndex, nil)
	if err != nil {
		return err
	}

	err = esd.dbClient.CheckAndCreateIndex(validatorsIndex, nil)
	if err != nil {
		return err
	}

	err = esd.dbClient.CheckAndCreateIndex(roundIndex, timestampMapping())
	if err != nil {
		return err
	}

	err = esd.dbClient.CheckAndCreateIndex(ratingIndex, nil)
	if err != nil {
		return err
	}

	err = esd.dbClient.CheckAndCreateIndex(miniblocksIndex, nil)
	if err != nil {
		return err
	}

	return nil
}

// SaveHeader will prepare and save information about a header in elasticsearch server
func (esd *elasticSearchDatabase) SaveHeader(
	header data.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	txsSize int,
) {
	var buff bytes.Buffer

	serializedBlock, headerHash := esd.getSerializedElasticBlockAndHeaderHash(header, signersIndexes, body, notarizedHeadersHashes, txsSize)

	buff.Grow(len(serializedBlock))
	_, err := buff.Write(serializedBlock)
	if err != nil {
		log.Warn("elastic search: save header, write", "error", err.Error())
	}

	req := &esapi.IndexRequest{
		Index:      blockIndex,
		DocumentID: hex.EncodeToString(headerHash),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	err = esd.dbClient.DoRequest(req)
	if err != nil {
		log.Warn("indexer: could not index block header", "error", err.Error())
		return
	}
}

func (esd *elasticSearchDatabase) getSerializedElasticBlockAndHeaderHash(
	header data.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	sizeTxs int,
) ([]byte, []byte) {
	headerBytes, err := esd.marshalizer.Marshal(header)
	if err != nil {
		log.Debug("indexer: marshal header", "error", err)
		return nil, nil
	}
	bodyBytes, err := esd.marshalizer.Marshal(body)
	if err != nil {
		log.Debug("indexer: marshal body", "error", err)
		return nil, nil
	}

	blockSizeInBytes := len(headerBytes) + len(bodyBytes)

	miniblocksHashes := make([]string, 0)
	for _, miniblock := range body.MiniBlocks {
		mbHash, errComputeHash := core.CalculateHash(esd.marshalizer, esd.hasher, miniblock)
		if errComputeHash != nil {
			log.Warn("internal error computing hash", "error", errComputeHash)

			continue
		}

		encodedMbHash := hex.EncodeToString(mbHash)
		miniblocksHashes = append(miniblocksHashes, encodedMbHash)
	}

	headerHash := esd.hasher.Compute(string(headerBytes))
	elasticBlock := Block{
		Nonce:                 header.GetNonce(),
		Round:                 header.GetRound(),
		Epoch:                 header.GetEpoch(),
		ShardID:               header.GetShardID(),
		Hash:                  hex.EncodeToString(headerHash),
		MiniBlocksHashes:      miniblocksHashes,
		NotarizedBlocksHashes: notarizedHeadersHashes,
		PubKeyBitmap:          hex.EncodeToString(header.GetPubKeysBitmap()),
		Size:                  int64(blockSizeInBytes),
		SizeTxs:               int64(sizeTxs),
		Timestamp:             time.Duration(header.GetTimeStamp()),
		TxCount:               header.GetTxCount(),
		StateRootHash:         hex.EncodeToString(header.GetRootHash()),
		PrevHash:              hex.EncodeToString(header.GetPrevHash()),
	}

	if header.GetNonce() == 0 {
		elasticBlock.PrevHash = hex.EncodeToString(header.GetReserved())
		elasticBlock.Proposer = 0
		elasticBlock.Validators = []uint64{0}
	} else {
		elasticBlock.Proposer = signersIndexes[0]
		elasticBlock.Validators = signersIndexes
	}

	serializedBlock, err := json.Marshal(elasticBlock)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal elastic header")
		return nil, nil
	}

	return serializedBlock, headerHash
}

// SaveTransactions will prepare and save information about a transactions in elasticsearch server
func (esd *elasticSearchDatabase) SaveTransactions(
	body *block.Body,
	header data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	selfShardID uint32,
	mbsHashInDB map[string]bool,
) {
	txs := esd.prepareTransactionsForDatabase(body, header, txPool, selfShardID)
	buffSlice := serializeTransactions(txs, selfShardID, esd.foundedObjMap, mbsHashInDB)

	for idx := range buffSlice {
		err := esd.dbClient.DoBulkRequest(&buffSlice[idx], txIndex)
		if err != nil {
			log.Warn("indexer indexing bulk of transactions",
				"error", err.Error())
			continue
		}
	}
}

// SetTxLogsProcessor will set tx logs processor
func (esd *elasticSearchDatabase) SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase) {
	esd.txLogsProcessor = txLogsProc
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (esd *elasticSearchDatabase) SaveMiniblocks(header data.HeaderHandler, body *block.Body) map[string]bool {
	miniblocks := esd.getMiniblocks(header, body)
	if miniblocks == nil {
		log.Warn("indexer: could not index miniblocks")
		return make(map[string]bool)
	}
	if len(miniblocks) == 0 {
		return make(map[string]bool)
	}

	buff, mbHashDb := serializeBulkMiniBlocks(header.GetShardID(), miniblocks, esd.foundedObjMap)
	err := esd.dbClient.DoBulkRequest(&buff, miniblocksIndex)
	if err != nil {
		log.Warn("indexing bulk of miniblocks", "error", err.Error())
	}

	return mbHashDb
}

func (esd *elasticSearchDatabase) getMiniblocks(header data.HeaderHandler, body *block.Body) []*Miniblock {
	headerHash, err := core.CalculateHash(esd.marshalizer, esd.hasher, header)
	if err != nil {
		log.Warn("indexer: could not calculate header hash", "error", err.Error())
		return nil
	}

	encodedHeaderHash := hex.EncodeToString(headerHash)

	miniblocks := make([]*Miniblock, 0)
	for _, miniblock := range body.MiniBlocks {
		mbHash, errComputeHash := core.CalculateHash(esd.marshalizer, esd.hasher, miniblock)
		if errComputeHash != nil {
			log.Warn("internal error computing hash", "error", errComputeHash)

			continue
		}

		encodedMbHash := hex.EncodeToString(mbHash)

		mb := &Miniblock{
			Hash:            encodedMbHash,
			SenderShardID:   miniblock.SenderShardID,
			ReceiverShardID: miniblock.ReceiverShardID,
			Type:            miniblock.Type.String(),
		}

		if mb.SenderShardID == header.GetShardID() {
			mb.SenderBlockHash = encodedHeaderHash
		} else {
			mb.ReceiverBlockHash = encodedHeaderHash
		}

		if mb.SenderShardID == mb.ReceiverShardID {
			mb.ReceiverBlockHash = encodedHeaderHash
		}

		miniblocks = append(miniblocks, mb)
	}

	return miniblocks
}

// SaveRoundsInfos will prepare and save information about a slice of rounds in elasticsearch server
func (esd *elasticSearchDatabase) SaveRoundsInfos(infos []RoundInfo) {
	var buff bytes.Buffer

	for _, info := range infos {
		serializedRoundInfo, meta := serializeRoundInfo(info)

		buff.Grow(len(meta) + len(serializedRoundInfo))
		_, err := buff.Write(meta)
		if err != nil {
			log.Warn("indexer: cannot write meta", "error", err.Error())
		}

		_, err = buff.Write(serializedRoundInfo)
		if err != nil {
			log.Warn("indexer: cannot write serialized round info", "error", err.Error())
		}
	}

	err := esd.dbClient.DoBulkRequest(&buff, roundIndex)
	if err != nil {
		log.Warn("indexer: cannot index rounds info", "error", err.Error())
		return
	}
}

func serializeRoundInfo(info RoundInfo) ([]byte, []byte) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%d_%d", "_type" : "%s" } }%s`,
		info.ShardId, info.Index, "_doc", "\n"))

	serializedInfo, err := json.Marshal(info)
	if err != nil {
		log.Debug("indexer: could not serialize round info, will skip indexing this round info")
		return nil, nil
	}
	// append a newline foreach element in the bulk we create
	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
}

// SaveShardValidatorsPubKeys will prepare and save information about a shard validators public keys in elasticsearch server
func (esd *elasticSearchDatabase) SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) {
	var buff bytes.Buffer

	shardValPubKeys := ValidatorsPublicKeys{
		PublicKeys: make([]string, 0, len(shardValidatorsPubKeys)),
	}
	for _, validatorPk := range shardValidatorsPubKeys {
		strValidatorPk := esd.validatorPubkeyConverter.Encode(validatorPk)
		shardValPubKeys.PublicKeys = append(shardValPubKeys.PublicKeys, strValidatorPk)
	}

	marshalizedValidatorPubKeys, err := json.Marshal(shardValPubKeys)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal validators public keys")
		return
	}

	buff.Grow(len(marshalizedValidatorPubKeys))
	_, err = buff.Write(marshalizedValidatorPubKeys)
	if err != nil {
		log.Warn("elastic search: save shard validators pub keys, write", "error", err.Error())
	}

	req := &esapi.IndexRequest{
		Index:      validatorsIndex,
		DocumentID: fmt.Sprintf("%d_%d", shardID, epoch),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	err = esd.dbClient.DoRequest(req)
	if err != nil {
		log.Warn("indexer: can not index validators pubkey", "error", err.Error())
		return
	}
}

// SaveValidatorsRating will save validators rating
func (esd *elasticSearchDatabase) SaveValidatorsRating(index string, validatorsRatingInfo []ValidatorRatingInfo) {
	var buff bytes.Buffer

	infosRating := ValidatorsRatingInfo{ValidatorsInfos: validatorsRatingInfo}

	marshalizedInfoRating, err := json.Marshal(&infosRating)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal validators rating")
		return
	}

	buff.Grow(len(marshalizedInfoRating))
	_, err = buff.Write(marshalizedInfoRating)
	if err != nil {
		log.Warn("elastic search: save validators rating, write", "error", err.Error())
	}

	req := &esapi.IndexRequest{
		Index:      ratingIndex,
		DocumentID: index,
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	err = esd.dbClient.DoRequest(req)
	if err != nil {
		log.Warn("indexer: can not index validators rating", "error", err.Error())
		return
	}
}

// SaveShardStatistics will prepare and save information about a shard statistics in elasticsearch server
func (esd *elasticSearchDatabase) SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) {
	buff := prepareGeneralInfo(tpsBenchmark)

	for _, shardInfo := range tpsBenchmark.ShardStatistics() {
		serializedShardInfo, serializedMetaInfo := serializeShardInfo(shardInfo)
		if serializedShardInfo == nil {
			continue
		}

		buff.Grow(len(serializedMetaInfo) + len(serializedShardInfo))
		_, err := buff.Write(serializedMetaInfo)
		if err != nil {
			log.Warn("elastic search: update TPS write meta", "error", err.Error())
		}
		_, err = buff.Write(serializedShardInfo)
		if err != nil {
			log.Warn("elastic search: update TPS write serialized data", "error", err.Error())
		}

		err = esd.dbClient.DoBulkRequest(&buff, tpsIndex)
		if err != nil {
			log.Warn("indexer: error indexing tps information", "error", err.Error())
			continue
		}
	}
}

func (esd *elasticSearchDatabase) foundedObjMap(hashes []string, index string) (map[string]bool, error) {
	if len(hashes) == 0 {
		return make(map[string]bool), nil
	}

	response, err := esd.dbClient.DoMultiGet(getDocumentsByIDsQuery(hashes), index)
	if err != nil {
		return nil, err
	}

	return getDecodedResponseMultiGet(response), nil
}
