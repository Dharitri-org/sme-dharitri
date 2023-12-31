package blockchain

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/statusHandler"
)

var _ data.ChainHandler = (*blockChain)(nil)

// blockChain holds the block information for the current shard.
//
// The BlockChain also holds pointers to the Genesis block header and the current block
type blockChain struct {
	*baseBlockChain
}

// NewBlockChain returns an initialized blockchain
func NewBlockChain() *blockChain {
	return &blockChain{
		baseBlockChain: &baseBlockChain{
			appStatusHandler: statusHandler.NewNilStatusHandler(),
		},
	}
}

// SetGenesisHeader sets the genesis block header pointer
func (bc *blockChain) SetGenesisHeader(genesisBlock data.HeaderHandler) error {
	if check.IfNil(genesisBlock) {
		bc.mut.Lock()
		bc.genesisHeader = nil
		bc.mut.Unlock()

		return nil
	}

	gb, ok := genesisBlock.(*block.Header)
	if !ok {
		return data.ErrInvalidHeaderType
	}
	bc.mut.Lock()
	bc.genesisHeader = gb.Clone()
	bc.mut.Unlock()

	return nil
}

// SetCurrentBlockHeader sets current block header pointer
func (bc *blockChain) SetCurrentBlockHeader(header data.HeaderHandler) error {
	if check.IfNil(header) {
		bc.mut.Lock()
		bc.currentBlockHeader = nil
		bc.mut.Unlock()

		return nil
	}

	h, ok := header.(*block.Header)
	if !ok {
		return data.ErrInvalidHeaderType
	}

	bc.appStatusHandler.SetUInt64Value(core.MetricNonce, h.Nonce)
	bc.appStatusHandler.SetUInt64Value(core.MetricSynchronizedRound, h.Round)

	bc.mut.Lock()
	bc.currentBlockHeader = h.Clone()
	bc.mut.Unlock()

	return nil
}

// CreateNewHeader creates a new header
func (bc *blockChain) CreateNewHeader() data.HeaderHandler {
	return &block.Header{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bc *blockChain) IsInterfaceNil() bool {
	return bc == nil || bc.baseBlockChain == nil
}
