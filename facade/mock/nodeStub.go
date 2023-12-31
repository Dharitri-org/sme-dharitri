package mock

import (
	"encoding/hex"
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/api/block"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/debug"
	"github.com/Dharitri-org/sme-dharitri/heartbeat/data"
)

// NodeStub -
type NodeStub struct {
	AddressHandler             func() (string, error)
	ConnectToAddressesHandler  func([]string) error
	StartConsensusHandler      func() error
	GetBalanceHandler          func(address string) (*big.Int, error)
	GenerateTransactionHandler func(sender string, receiver string, amount string, code string) (*transaction.Transaction, error)
	CreateTransactionHandler   func(nonce uint64, value string, receiverHex string, senderHex string, gasPrice uint64,
		gasLimit uint64, data []byte, signatureHex string, chainID string, version uint32) (*transaction.Transaction, []byte, error)
	ValidateTransactionHandler                     func(tx *transaction.Transaction) error
	GetTransactionHandler                          func(hash string) (*transaction.ApiTransactionResult, error)
	SendBulkTransactionsHandler                    func(txs []*transaction.Transaction) (uint64, error)
	GetAccountHandler                              func(address string) (state.UserAccountHandler, error)
	GetCurrentPublicKeyHandler                     func() string
	GenerateAndSendBulkTransactionsHandler         func(destination string, value *big.Int, nrTransactions uint64) error
	GenerateAndSendBulkTransactionsOneByOneHandler func(destination string, value *big.Int, nrTransactions uint64) error
	GetHeartbeatsHandler                           func() []data.PubKeyHeartbeat
	ValidatorStatisticsApiCalled                   func() (map[string]*state.ValidatorApiResponse, error)
	DirectTriggerCalled                            func(epoch uint32) error
	IsSelfTriggerCalled                            func() bool
	GetQueryHandlerCalled                          func(name string) (debug.QueryHandler, error)
	GetValueForKeyCalled                           func(address string, key string) (string, error)
	GetPeerInfoCalled                              func(pid string) ([]core.QueryP2PPeerInfo, error)
	GetBlockByHashCalled                           func(hash string, withTxs bool) (*block.APIBlock, error)
	GetBlockByNonceCalled                          func(nonce uint64, withTxs bool) (*block.APIBlock, error)
}

// GetValueForKey -
func (ns *NodeStub) GetValueForKey(address string, key string) (string, error) {
	if ns.GetValueForKeyCalled != nil {
		return ns.GetValueForKeyCalled(address, key)
	}

	return "", nil
}

// EncodeAddressPubkey -
func (ns *NodeStub) EncodeAddressPubkey(pk []byte) (string, error) {
	return hex.EncodeToString(pk), nil
}

// GetBlockByHash -
func (ns *NodeStub) GetBlockByHash(hash string, withTxs bool) (*block.APIBlock, error) {
	return ns.GetBlockByHashCalled(hash, withTxs)
}

// GetBlockByNonce -
func (ns *NodeStub) GetBlockByNonce(nonce uint64, withTxs bool) (*block.APIBlock, error) {
	return ns.GetBlockByNonceCalled(nonce, withTxs)
}

// DecodeAddressPubkey -
func (ns *NodeStub) DecodeAddressPubkey(pk string) ([]byte, error) {
	return hex.DecodeString(pk)
}

// StartConsensus -
func (ns *NodeStub) StartConsensus() error {
	return ns.StartConsensusHandler()
}

// GetBalance -
func (ns *NodeStub) GetBalance(address string) (*big.Int, error) {
	return ns.GetBalanceHandler(address)
}

// CreateTransaction -
func (ns *NodeStub) CreateTransaction(nonce uint64, value string, receiverHex string, senderHex string, gasPrice uint64,
	gasLimit uint64, data []byte, signatureHex string, chainID string, version uint32) (*transaction.Transaction, []byte, error) {

	return ns.CreateTransactionHandler(nonce, value, receiverHex, senderHex, gasPrice, gasLimit, data, signatureHex, chainID, version)
}

// ValidateTransaction --
func (ns *NodeStub) ValidateTransaction(tx *transaction.Transaction) error {
	return ns.ValidateTransactionHandler(tx)
}

// GetTransaction -
func (ns *NodeStub) GetTransaction(hash string) (*transaction.ApiTransactionResult, error) {
	return ns.GetTransactionHandler(hash)
}

// SendBulkTransactions -
func (ns *NodeStub) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return ns.SendBulkTransactionsHandler(txs)
}

// GetAccount -
func (ns *NodeStub) GetAccount(address string) (state.UserAccountHandler, error) {
	return ns.GetAccountHandler(address)
}

// GetHeartbeats -
func (ns *NodeStub) GetHeartbeats() []data.PubKeyHeartbeat {
	return ns.GetHeartbeatsHandler()
}

// ValidatorStatisticsApi -
func (ns *NodeStub) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return ns.ValidatorStatisticsApiCalled()
}

// DirectTrigger -
func (ns *NodeStub) DirectTrigger(epoch uint32) error {
	return ns.DirectTriggerCalled(epoch)
}

// IsSelfTrigger -
func (ns *NodeStub) IsSelfTrigger() bool {
	return ns.IsSelfTriggerCalled()
}

// GetQueryHandler -
func (ns *NodeStub) GetQueryHandler(name string) (debug.QueryHandler, error) {
	if ns.GetQueryHandlerCalled != nil {
		return ns.GetQueryHandlerCalled(name)
	}

	return nil, nil
}

// GetPeerInfo -
func (ns *NodeStub) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	if ns.GetPeerInfoCalled != nil {
		return ns.GetPeerInfoCalled(pid)
	}

	return make([]core.QueryP2PPeerInfo, 0), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ns *NodeStub) IsInterfaceNil() bool {
	return ns == nil
}
