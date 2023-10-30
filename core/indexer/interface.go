package indexer

import (
	"bytes"
	"io"

	"github.com/Dharitri-org/sme-dharitri/core/statistics"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// Indexer is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Indexer interface {
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
	SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler, signersIndexes []uint64, notarizedHeadersHashes []string)
	SaveRoundsInfos(roundsInfos []RoundInfo)
	UpdateTPS(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRating(indexID string, infoRating []ValidatorRatingInfo)
	IsInterfaceNil() bool
	IsNilIndexer() bool
}

// databaseHandler is an interface used by elasticsearch component to prepare data to be saved on elasticseach server
type databaseHandler interface {
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
	SaveHeader(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int)
	SaveMiniblocks(header data.HeaderHandler, body *block.Body) map[string]bool
	SaveTransactions(body *block.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool)
	SaveRoundsInfos(infos []RoundInfo)
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte)
	SaveValidatorsRating(Index string, validatorsRatingInfo []ValidatorRatingInfo)
	SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark)
}

// databaseClientHandler is an interface that do requests to elasticsearch server
type databaseClientHandler interface {
	DoRequest(req *esapi.IndexRequest) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoMultiGet(query object, index string) (object, error)
	CheckAndCreateIndex(index string, body io.Reader) error
}
