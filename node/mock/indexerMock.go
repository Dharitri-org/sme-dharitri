package mock

import (
	"github.com/Dharitri-org/sme-dharitri/core/indexer"
	"github.com/Dharitri-org/sme-dharitri/core/statistics"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/process"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct {
	SaveBlockCalled func(body *block.Body, header *block.Header)
}

// SaveBlock -
func (im *IndexerMock) SaveBlock(_ data.BodyHandler, _ data.HeaderHandler, _ map[string]data.TransactionHandler, _ []uint64, _ []string) {
	panic("implement me")
}

// SetTxLogsProcessor will do nothing
func (im *IndexerMock) SetTxLogsProcessor(_ process.TransactionLogProcessorDatabase) {
}

// UpdateTPS -
func (im *IndexerMock) UpdateTPS(_ statistics.TPSBenchmark) {
	panic("implement me")
}

// SaveRoundsInfos -
func (im *IndexerMock) SaveRoundsInfos(_ []indexer.RoundInfo) {
	panic("implement me")
}

// SaveValidatorsRating --
func (im *IndexerMock) SaveValidatorsRating(_ string, _ []indexer.ValidatorRatingInfo) {

}

// SaveValidatorsPubKeys -
func (im *IndexerMock) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (im *IndexerMock) IsInterfaceNil() bool {
	return im == nil
}

// IsNilIndexer -
func (im *IndexerMock) IsNilIndexer() bool {
	return false
}
