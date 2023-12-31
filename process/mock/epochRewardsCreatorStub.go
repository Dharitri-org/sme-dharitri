package mock

import (
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/state"
)

// EpochRewardsCreatorStub -
type EpochRewardsCreatorStub struct {
	CreateRewardsMiniBlocksCalled  func(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error)
	VerifyRewardsMiniBlocksCalled  func(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error
	CreateMarshalizedDataCalled    func(body *block.Body) map[string][][]byte
	SaveTxBlockToStorageCalled     func(metaBlock *block.MetaBlock, body *block.Body)
	DeleteTxsFromStorageCalled     func(metaBlock *block.MetaBlock, body *block.Body)
	RemoveBlockDataFromPoolsCalled func(metaBlock *block.MetaBlock, body *block.Body)
	GetRewardsTxsCalled            func(body *block.Body) map[string]data.TransactionHandler
	GetProtocolSustainCalled       func() *big.Int
}

// GetProtocolSustainabilityRewards -
func (e *EpochRewardsCreatorStub) GetProtocolSustainabilityRewards() *big.Int {
	if e.GetProtocolSustainCalled != nil {
		return e.GetProtocolSustainCalled()
	}
	return big.NewInt(0)
}

// CreateRewardsMiniBlocks -
func (e *EpochRewardsCreatorStub) CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	if e.CreateRewardsMiniBlocksCalled != nil {
		return e.CreateRewardsMiniBlocksCalled(metaBlock, validatorsInfo)
	}
	return nil, nil
}

// VerifyRewardsMiniBlocks -
func (e *EpochRewardsCreatorStub) VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error {
	if e.VerifyRewardsMiniBlocksCalled != nil {
		return e.VerifyRewardsMiniBlocksCalled(metaBlock, validatorsInfo)
	}
	return nil
}

// CreateMarshalizedData -
func (e *EpochRewardsCreatorStub) CreateMarshalizedData(body *block.Body) map[string][][]byte {
	if e.CreateMarshalizedDataCalled != nil {
		return e.CreateMarshalizedDataCalled(body)
	}
	return nil
}

// GetRewardsTxs --
func (e *EpochRewardsCreatorStub) GetRewardsTxs(body *block.Body) map[string]data.TransactionHandler {
	if e.GetRewardsTxsCalled != nil {
		return e.GetRewardsTxsCalled(body)
	}
	return nil
}

// SaveTxBlockToStorage -
func (e *EpochRewardsCreatorStub) SaveTxBlockToStorage(metaBlock *block.MetaBlock, body *block.Body) {
	if e.SaveTxBlockToStorageCalled != nil {
		e.SaveTxBlockToStorageCalled(metaBlock, body)
	}
}

// DeleteTxsFromStorage -
func (e *EpochRewardsCreatorStub) DeleteTxsFromStorage(metaBlock *block.MetaBlock, body *block.Body) {
	if e.DeleteTxsFromStorageCalled != nil {
		e.DeleteTxsFromStorageCalled(metaBlock, body)
	}
}

// IsInterfaceNil -
func (e *EpochRewardsCreatorStub) IsInterfaceNil() bool {
	return e == nil
}

// RemoveBlockDataFromPools -
func (e *EpochRewardsCreatorStub) RemoveBlockDataFromPools(metaBlock *block.MetaBlock, body *block.Body) {
	if e.RemoveBlockDataFromPoolsCalled != nil {
		e.RemoveBlockDataFromPoolsCalled(metaBlock, body)
	}
}
