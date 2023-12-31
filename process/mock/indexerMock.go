package mock

import (
	"github.com/Dharitri-org/sme-dharitri/core/indexer"
	"github.com/Dharitri-org/sme-dharitri/core/statistics"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/process"
)

// IndexerMock is a mock implementation fot the Indexer interface
type IndexerMock struct {
	SaveBlockCalled func(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler)
}

// SaveBlock -
func (im *IndexerMock) SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler, _ []uint64, _ []string) {
	if im.SaveBlockCalled != nil {
		im.SaveBlockCalled(body, header, txPool)
	}
}

// SetTxLogsProcessor will do nothing
func (im *IndexerMock) SetTxLogsProcessor(_ process.TransactionLogProcessorDatabase) {
}

// SaveValidatorsRating --
func (im *IndexerMock) SaveValidatorsRating(_ string, _ []indexer.ValidatorRatingInfo) {

}

// SaveMetaBlock -
func (im *IndexerMock) SaveMetaBlock(_ data.HeaderHandler, _ []uint64) {
}

// UpdateTPS -
func (im *IndexerMock) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveRoundsInfos -
func (im *IndexerMock) SaveRoundsInfos(_ []indexer.RoundInfo) {
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
