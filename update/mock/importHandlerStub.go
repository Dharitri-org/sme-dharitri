package mock

import (
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/state"
)

// ImportHandlerStub -
type ImportHandlerStub struct {
	ImportAllCalled              func() error
	GetValidatorAccountsDBCalled func() state.AccountsAdapter
	GetMiniBlocksCalled          func() map[string]*block.MiniBlock
	GetHardForkMetaBlockCalled   func() *block.MetaBlock
	GetTransactionsCalled        func() map[string]data.TransactionHandler
	GetAccountsDBForShardCalled  func(shardID uint32) state.AccountsAdapter
}

// ImportAll -
func (ihs *ImportHandlerStub) ImportAll() error {
	if ihs.ImportAllCalled != nil {
		return ihs.ImportAllCalled()
	}
	return nil
}

// GetValidatorAccountsDB -
func (ihs *ImportHandlerStub) GetValidatorAccountsDB() state.AccountsAdapter {
	if ihs.GetValidatorAccountsDBCalled != nil {
		return ihs.GetValidatorAccountsDBCalled()
	}
	return nil
}

// GetMiniBlocks -
func (ihs *ImportHandlerStub) GetMiniBlocks() map[string]*block.MiniBlock {
	if ihs.GetMiniBlocksCalled != nil {
		return ihs.GetMiniBlocksCalled()
	}
	return nil
}

// GetHardForkMetaBlock -
func (ihs *ImportHandlerStub) GetHardForkMetaBlock() *block.MetaBlock {
	if ihs.GetHardForkMetaBlockCalled != nil {
		return ihs.GetHardForkMetaBlockCalled()
	}
	return nil
}

// GetTransactions -
func (ihs *ImportHandlerStub) GetTransactions() map[string]data.TransactionHandler {
	if ihs.GetTransactionsCalled != nil {
		return ihs.GetTransactionsCalled()
	}
	return nil
}

// GetAccountsDBForShard -
func (ihs *ImportHandlerStub) GetAccountsDBForShard(shardID uint32) state.AccountsAdapter {
	if ihs.GetAccountsDBForShardCalled != nil {
		return ihs.GetAccountsDBForShardCalled(shardID)
	}
	return nil
}

// IsInterfaceNil -
func (ihs *ImportHandlerStub) IsInterfaceNil() bool {
	return ihs == nil
}
