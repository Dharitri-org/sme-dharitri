package facade

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/Dharitri-org/sme-dharitri/api"
	"github.com/Dharitri-org/sme-dharitri/api/address"
	"github.com/Dharitri-org/sme-dharitri/api/block"
	"github.com/Dharitri-org/sme-dharitri/api/hardfork"
	"github.com/Dharitri-org/sme-dharitri/api/middleware"
	"github.com/Dharitri-org/sme-dharitri/api/node"
	transactionApi "github.com/Dharitri-org/sme-dharitri/api/transaction"
	"github.com/Dharitri-org/sme-dharitri/api/validator"
	"github.com/Dharitri-org/sme-dharitri/api/vmValues"
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/core/statistics"
	"github.com/Dharitri-org/sme-dharitri/core/throttler"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/debug"
	"github.com/Dharitri-org/sme-dharitri/heartbeat/data"
	"github.com/Dharitri-org/sme-dharitri/node/external"
	"github.com/Dharitri-org/sme-dharitri/ntp"
	"github.com/Dharitri-org/sme-dharitri/process"
	logger "github.com/Dharitri-org/sme-logger"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

// DefaultRestInterface is the default interface the rest API will start on if not specified
const DefaultRestInterface = "localhost:8080"

// DefaultRestPortOff is the default value that should be passed if it is desired
//
//	to start the node without a REST endpoint available
const DefaultRestPortOff = "off"

var _ = address.FacadeHandler(&nodeFacade{})
var _ = hardfork.FacadeHandler(&nodeFacade{})
var _ = node.FacadeHandler(&nodeFacade{})
var _ = transactionApi.FacadeHandler(&nodeFacade{})
var _ = validator.FacadeHandler(&nodeFacade{})
var _ = vmValues.FacadeHandler(&nodeFacade{})

var log = logger.GetOrCreate("facade")

type resetHandler interface {
	Reset()
	IsInterfaceNil() bool
}

// ArgNodeFacade represents the argument for the nodeFacade
type ArgNodeFacade struct {
	Node                   NodeHandler
	ApiResolver            ApiResolver
	RestAPIServerDebugMode bool
	WsAntifloodConfig      config.WebServerAntifloodConfig
	FacadeConfig           config.FacadeConfig
	ApiRoutesConfig        config.ApiRoutesConfig
	AccountsState          state.AccountsAdapter
	PeerState              state.AccountsAdapter
}

// nodeFacade represents a facade for grouping the functionality for the node
type nodeFacade struct {
	node                   NodeHandler
	apiResolver            ApiResolver
	syncer                 ntp.SyncTimer
	tpsBenchmark           *statistics.TpsBenchmark
	config                 config.FacadeConfig
	apiRoutesConfig        config.ApiRoutesConfig
	endpointsThrottlers    map[string]core.Throttler
	wsAntifloodConfig      config.WebServerAntifloodConfig
	restAPIServerDebugMode bool
	accountsState          state.AccountsAdapter
	peerState              state.AccountsAdapter
	ctx                    context.Context
	cancelFunc             func()
}

// NewNodeFacade creates a new Facade with a NodeWrapper
func NewNodeFacade(arg ArgNodeFacade) (*nodeFacade, error) {
	if check.IfNil(arg.Node) {
		return nil, ErrNilNode
	}
	if check.IfNil(arg.ApiResolver) {
		return nil, ErrNilApiResolver
	}
	if len(arg.ApiRoutesConfig.APIPackages) == 0 {
		return nil, ErrNoApiRoutesConfig
	}
	if arg.WsAntifloodConfig.SimultaneousRequests == 0 {
		return nil, fmt.Errorf("%w, SimultaneousRequests should not be 0", ErrInvalidValue)
	}
	if arg.WsAntifloodConfig.SameSourceRequests == 0 {
		return nil, fmt.Errorf("%w, SameSourceRequests should not be 0", ErrInvalidValue)
	}
	if arg.WsAntifloodConfig.SameSourceResetIntervalInSec == 0 {
		return nil, fmt.Errorf("%w, SameSourceResetIntervalInSec should not be 0", ErrInvalidValue)
	}
	if check.IfNil(arg.AccountsState) {
		return nil, ErrNilAccountState
	}
	if check.IfNil(arg.PeerState) {
		return nil, ErrNilPeerState
	}

	throttlersMap := computeEndpointsNumGoRoutinesThrottlers(arg.WsAntifloodConfig)

	nf := &nodeFacade{
		node:                   arg.Node,
		apiResolver:            arg.ApiResolver,
		restAPIServerDebugMode: arg.RestAPIServerDebugMode,
		wsAntifloodConfig:      arg.WsAntifloodConfig,
		config:                 arg.FacadeConfig,
		apiRoutesConfig:        arg.ApiRoutesConfig,
		endpointsThrottlers:    throttlersMap,
		accountsState:          arg.AccountsState,
		peerState:              arg.PeerState,
	}
	nf.ctx, nf.cancelFunc = context.WithCancel(context.Background())

	return nf, nil
}

func computeEndpointsNumGoRoutinesThrottlers(webServerAntiFloodConfig config.WebServerAntifloodConfig) map[string]core.Throttler {
	throttlersMap := make(map[string]core.Throttler)
	for _, endpointSetting := range webServerAntiFloodConfig.EndpointsThrottlers {
		newThrottler, err := throttler.NewNumGoRoutinesThrottler(endpointSetting.MaxNumGoRoutines)
		if err != nil {
			log.Warn("error when setting the maximum go routines throttler for endpoint",
				"endpoint", endpointSetting.Endpoint,
				"max go routines", endpointSetting.MaxNumGoRoutines,
				"error", err,
			)
			continue
		}
		throttlersMap[endpointSetting.Endpoint] = newThrottler
	}

	return throttlersMap
}

// SetSyncer sets the current syncer
func (nf *nodeFacade) SetSyncer(syncer ntp.SyncTimer) {
	nf.syncer = syncer
}

// SetTpsBenchmark sets the tps benchmark handler
func (nf *nodeFacade) SetTpsBenchmark(tpsBenchmark *statistics.TpsBenchmark) {
	nf.tpsBenchmark = tpsBenchmark
}

// TpsBenchmark returns the tps benchmark handler
func (nf *nodeFacade) TpsBenchmark() *statistics.TpsBenchmark {
	return nf.tpsBenchmark
}

// StartNode starts the underlying node
func (nf *nodeFacade) StartNode() error {
	return nf.node.StartConsensus()
}

// StartBackgroundServices starts all background services needed for the correct functionality of the node
func (nf *nodeFacade) StartBackgroundServices() {
	go nf.startRest()
}

// RestAPIServerDebugMode return true is debug mode for Rest API is enabled
func (nf *nodeFacade) RestAPIServerDebugMode() bool {
	return nf.restAPIServerDebugMode
}

// RestApiInterface returns the interface on which the rest API should start on, based on the config file provided.
// The API will start on the DefaultRestInterface value unless a correct value is passed or
//
//	the value is explicitly set to off, in which case it will not start at all
func (nf *nodeFacade) RestApiInterface() string {
	if nf.config.RestApiInterface == "" {
		return DefaultRestInterface
	}

	return nf.config.RestApiInterface
}

func (nf *nodeFacade) startRest() {
	log.Trace("starting REST api server")

	switch nf.RestApiInterface() {
	case DefaultRestPortOff:
		log.Debug("web server is off")
	default:
		log.Debug("creating web server limiters")
		limiters, err := nf.createMiddlewareLimiters()
		if err != nil {
			log.Error("error creating web server limiters",
				"error", err.Error(),
			)
			log.Error("web server is off")
			return
		}

		log.Debug("starting web server",
			"SimultaneousRequests", nf.wsAntifloodConfig.SimultaneousRequests,
			"SameSourceRequests", nf.wsAntifloodConfig.SameSourceRequests,
			"SameSourceResetIntervalInSec", nf.wsAntifloodConfig.SameSourceResetIntervalInSec,
		)

		err = api.Start(nf, nf.apiRoutesConfig, limiters...)
		if err != nil {
			log.Error("could not start webserver",
				"error", err.Error(),
			)
		}
	}
}

func (nf *nodeFacade) createMiddlewareLimiters() ([]api.MiddlewareProcessor, error) {
	sourceLimiter, err := middleware.NewSourceThrottler(nf.wsAntifloodConfig.SameSourceRequests)
	if err != nil {
		return nil, err
	}
	go nf.sourceLimiterReset(sourceLimiter)

	globalLimiter, err := middleware.NewGlobalThrottler(nf.wsAntifloodConfig.SimultaneousRequests)
	if err != nil {
		return nil, err
	}

	return []api.MiddlewareProcessor{sourceLimiter, globalLimiter}, nil
}

func (nf *nodeFacade) sourceLimiterReset(reset resetHandler) {
	betweenResetDuration := time.Second * time.Duration(nf.wsAntifloodConfig.SameSourceResetIntervalInSec)
	for {
		select {
		case <-time.After(betweenResetDuration):
			log.Trace("calling reset on WS source limiter")
			reset.Reset()
		case <-nf.ctx.Done():
			log.Debug("closing nodeFacade.sourceLimiterReset go routine")
			return
		}
	}
}

// GetBalance gets the current balance for a specified address
func (nf *nodeFacade) GetBalance(address string) (*big.Int, error) {
	return nf.node.GetBalance(address)
}

// GetValueForKey gets the value for a key in a given address
func (nf *nodeFacade) GetValueForKey(address string, key string) (string, error) {
	return nf.node.GetValueForKey(address, key)
}

// CreateTransaction creates a transaction from all needed fields
func (nf *nodeFacade) CreateTransaction(
	nonce uint64,
	value string,
	receiverHex string,
	senderHex string,
	gasPrice uint64,
	gasLimit uint64,
	txData []byte,
	signatureHex string,
	chainID string,
	version uint32,
) (*transaction.Transaction, []byte, error) {

	return nf.node.CreateTransaction(nonce, value, receiverHex, senderHex, gasPrice, gasLimit, txData, signatureHex, chainID, version)
}

// ValidateTransaction will validate a transaction
func (nf *nodeFacade) ValidateTransaction(tx *transaction.Transaction) error {
	return nf.node.ValidateTransaction(tx)
}

// ValidatorStatisticsApi will return the statistics for all validators
func (nf *nodeFacade) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return nf.node.ValidatorStatisticsApi()
}

// SendBulkTransactions will send a bulk of transactions on the topic channel
func (nf *nodeFacade) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return nf.node.SendBulkTransactions(txs)
}

// GetTransaction gets the transaction with a specified hash
func (nf *nodeFacade) GetTransaction(hash string) (*transaction.ApiTransactionResult, error) {
	return nf.node.GetTransaction(hash)
}

// ComputeTransactionGasLimit will estimate how many gas a transaction will consume
func (nf *nodeFacade) ComputeTransactionGasLimit(tx *transaction.Transaction) (uint64, error) {
	return nf.apiResolver.ComputeTransactionGasLimit(tx)
}

// GetAccount returns an accountResponse containing information
// about the account correlated with provided address
func (nf *nodeFacade) GetAccount(address string) (state.UserAccountHandler, error) {
	return nf.node.GetAccount(address)
}

// GetHeartbeats returns the heartbeat status for each public key from initial list or later joined to the network
func (nf *nodeFacade) GetHeartbeats() ([]data.PubKeyHeartbeat, error) {
	hbStatus := nf.node.GetHeartbeats()
	if hbStatus == nil {
		return nil, ErrHeartbeatsNotActive
	}

	return hbStatus, nil
}

// StatusMetrics will return the node's status metrics
func (nf *nodeFacade) StatusMetrics() external.StatusMetricsHandler {
	return nf.apiResolver.StatusMetrics()
}

// ExecuteSCQuery retrieves data from existing SC trie
func (nf *nodeFacade) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return nf.apiResolver.ExecuteSCQuery(query)
}

// PprofEnabled returns if profiling mode should be active or not on the application
func (nf *nodeFacade) PprofEnabled() bool {
	return nf.config.PprofEnabled
}

// Trigger will trigger a hardfork event
func (nf *nodeFacade) Trigger(epoch uint32) error {
	return nf.node.DirectTrigger(epoch)
}

// IsSelfTrigger returns true if the self public key is the same with the registered public key
func (nf *nodeFacade) IsSelfTrigger() bool {
	return nf.node.IsSelfTrigger()
}

// EncodeAddressPubkey will encode the provided address public key bytes to string
func (nf *nodeFacade) EncodeAddressPubkey(pk []byte) (string, error) {
	return nf.node.EncodeAddressPubkey(pk)
}

// DecodeAddressPubkey will try to decode the provided address public key string
func (nf *nodeFacade) DecodeAddressPubkey(pk string) ([]byte, error) {
	return nf.node.DecodeAddressPubkey(pk)
}

// GetQueryHandler returns the query handler if existing
func (nf *nodeFacade) GetQueryHandler(name string) (debug.QueryHandler, error) {
	return nf.node.GetQueryHandler(name)
}

// GetPeerInfo returns the peer info of a provided pid
func (nf *nodeFacade) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	return nf.node.GetPeerInfo(pid)
}

// GetThrottlerForEndpoint returns the throttler for a given endpoint if found
func (nf *nodeFacade) GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool) {
	throttlerForEndpoint, ok := nf.endpointsThrottlers[endpoint]
	isThrottlerOk := ok && throttlerForEndpoint != nil

	return throttlerForEndpoint, isThrottlerOk
}

// GetBlockByHash return the block for a given hash
func (nf *nodeFacade) GetBlockByHash(hash string, withTxs bool) (*block.APIBlock, error) {
	return nf.node.GetBlockByHash(hash, withTxs)
}

// GetBlockByNonce returns the block for a given nonce
func (nf *nodeFacade) GetBlockByNonce(nonce uint64, withTxs bool) (*block.APIBlock, error) {
	return nf.node.GetBlockByNonce(nonce, withTxs)
}

// Close will cleanup started go routines
// TODO use this close method
func (nf *nodeFacade) Close() error {
	nf.cancelFunc()

	return nil
}

// GetNumCheckpointsFromAccountState returns the number of checkpoints of the account state
func (nf *nodeFacade) GetNumCheckpointsFromAccountState() uint32 {
	return nf.accountsState.GetNumCheckpoints()
}

// GetNumCheckpointsFromPeerState returns the number of checkpoints of the peer state
func (nf *nodeFacade) GetNumCheckpointsFromPeerState() uint32 {
	return nf.peerState.GetNumCheckpoints()
}

// IsInterfaceNil returns true if there is no value under the interface
func (nf *nodeFacade) IsInterfaceNil() bool {
	return nf == nil
}
