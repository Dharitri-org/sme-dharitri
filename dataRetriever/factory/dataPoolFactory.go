package factory

import (
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/dataPool"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/dataPool/headersCache"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/shardedData"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/txpool"
	"github.com/Dharitri-org/sme-dharitri/process/economics"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/storage/factory"
	"github.com/Dharitri-org/sme-dharitri/storage/storageUnit"
	logger "github.com/Dharitri-org/sme-logger"
)

var log = logger.GetOrCreate("dataRetriever/factory")

// ArgsDataPool holds the arguments needed for NewDataPoolFromConfig function
type ArgsDataPool struct {
	Config           *config.Config
	EconomicsData    *economics.EconomicsData
	ShardCoordinator sharding.Coordinator
}

// NewDataPoolFromConfig will return a new instance of a PoolsHolder
func NewDataPoolFromConfig(args ArgsDataPool) (dataRetriever.PoolsHolder, error) {
	log.Debug("creatingDataPool from config")

	if args.Config == nil {
		return nil, dataRetriever.ErrNilConfig
	}
	if args.EconomicsData == nil {
		return nil, dataRetriever.ErrNilEconomicsData
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, dataRetriever.ErrNilShardCoordinator
	}

	mainConfig := args.Config

	txPool, err := txpool.NewShardedTxPool(txpool.ArgShardedTxPool{
		Config:         factory.GetCacherFromConfig(mainConfig.TxDataPool),
		MinGasPrice:    args.EconomicsData.MinGasPrice(),
		NumberOfShards: args.ShardCoordinator.NumberOfShards(),
		SelfShardID:    args.ShardCoordinator.SelfId(),
	})
	if err != nil {
		log.Error("error creating txpool")
		return nil, err
	}

	uTxPool, err := shardedData.NewShardedData(dataRetriever.UnsignedTxPoolName, factory.GetCacherFromConfig(mainConfig.UnsignedTransactionDataPool))
	if err != nil {
		log.Error("error creating smart contract result pool")
		return nil, err
	}

	rewardTxPool, err := shardedData.NewShardedData(dataRetriever.RewardTxPoolName, factory.GetCacherFromConfig(mainConfig.RewardTransactionDataPool))
	if err != nil {
		log.Error("error creating reward transaction pool")
		return nil, err
	}

	hdrPool, err := headersCache.NewHeadersPool(mainConfig.HeadersPoolConfig)
	if err != nil {
		log.Error("error creating headers pool")
		return nil, err
	}

	cacherCfg := factory.GetCacherFromConfig(mainConfig.TxBlockBodyDataPool)
	txBlockBody, err := storageUnit.NewCache(cacherCfg)
	if err != nil {
		log.Error("error creating txBlockBody")
		return nil, err
	}

	cacherCfg = factory.GetCacherFromConfig(mainConfig.PeerBlockBodyDataPool)
	peerChangeBlockBody, err := storageUnit.NewCache(cacherCfg)
	if err != nil {
		log.Error("error creating peerChangeBlockBody")
		return nil, err
	}

	cacherCfg = factory.GetCacherFromConfig(mainConfig.TrieNodesDataPool)
	trieNodes, err := storageUnit.NewCache(cacherCfg)
	if err != nil {
		log.Info("error creating trieNodes")
		return nil, err
	}

	currBlockTxs, err := dataPool.NewCurrentBlockPool()
	if err != nil {
		return nil, err
	}

	return dataPool.NewDataPool(
		txPool,
		uTxPool,
		rewardTxPool,
		hdrPool,
		txBlockBody,
		peerChangeBlockBody,
		trieNodes,
		currBlockTxs,
	)
}
