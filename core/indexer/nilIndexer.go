package indexer

import (
	"github.com/Dharitri-org/sme-dharitri/core/statistics"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/process"
)

// NilIndexer will be used when an Indexer is required, but another one isn't necessary or available
type NilIndexer struct {
}

// NewNilIndexer will return a Nil indexer
func NewNilIndexer() *NilIndexer {
	return new(NilIndexer)
}

// SaveBlock will do nothing
func (ni *NilIndexer) SaveBlock(_ data.BodyHandler, _ data.HeaderHandler, _ map[string]data.TransactionHandler, _ []uint64, _ []string) {
}

// SetTxLogsProcessor will do nothing
func (ni *NilIndexer) SetTxLogsProcessor(_ process.TransactionLogProcessorDatabase) {
}

// SaveRoundsInfos will do nothing
func (ni *NilIndexer) SaveRoundsInfos(_ []RoundInfo) {
}

// UpdateTPS will do nothing
func (ni *NilIndexer) UpdateTPS(_ statistics.TPSBenchmark) {
}

// SaveValidatorsRating --
func (ni *NilIndexer) SaveValidatorsRating(_ string, _ []ValidatorRatingInfo) {
}

// SaveValidatorsPubKeys will do nothing
func (ni *NilIndexer) SaveValidatorsPubKeys(_ map[uint32][][]byte, _ uint32) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (ni *NilIndexer) IsInterfaceNil() bool {
	return ni == nil
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (ni *NilIndexer) IsNilIndexer() bool {
	return true
}
