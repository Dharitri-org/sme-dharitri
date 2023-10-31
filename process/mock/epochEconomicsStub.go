package mock

import (
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/data/block"
)

// EpochEconomicsStub -
type EpochEconomicsStub struct {
	ComputeEndOfEpochEconomicsCalled func(metaBlock *block.MetaBlock) (*block.Economics, error)
	VerifyRewardsPerBlockCalled      func(metaBlock *block.MetaBlock, correctedProtocolSustainability *big.Int) error
}

// ComputeEndOfEpochEconomics -
func (e *EpochEconomicsStub) ComputeEndOfEpochEconomics(metaBlock *block.MetaBlock) (*block.Economics, error) {
	if e.ComputeEndOfEpochEconomicsCalled != nil {
		return e.ComputeEndOfEpochEconomicsCalled(metaBlock)
	}
	return &block.Economics{}, nil
}

// VerifyRewardsPerBlock -
func (e *EpochEconomicsStub) VerifyRewardsPerBlock(metaBlock *block.MetaBlock, correctedProtocolSustainability *big.Int) error {
	if e.VerifyRewardsPerBlockCalled != nil {
		return e.VerifyRewardsPerBlockCalled(metaBlock, correctedProtocolSustainability)
	}
	return nil
}

// IsInterfaceNil -
func (e *EpochEconomicsStub) IsInterfaceNil() bool {
	return e == nil
}
