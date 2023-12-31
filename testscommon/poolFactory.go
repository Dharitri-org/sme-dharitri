package testscommon

import (
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/dataPool"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/dataPool/headersCache"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/shardedData"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever/txpool"
	"github.com/Dharitri-org/sme-dharitri/storage/storageUnit"
)

// CreateTxPool -
func CreateTxPool(numShards uint32, selfShard uint32) (dataRetriever.ShardedDataCacherNotifier, error) {
	return txpool.NewShardedTxPool(
		txpool.ArgShardedTxPool{
			Config: storageUnit.CacheConfig{
				Capacity:             100_000,
				SizePerSender:        1_000_000_000,
				SizeInBytes:          1_000_000_000,
				SizeInBytesPerSender: 33_554_432,
				Shards:               16,
			},
			MinGasPrice:    200000000000,
			NumberOfShards: numShards,
			SelfShardID:    selfShard,
		},
	)
}

// CreatePoolsHolder -
func CreatePoolsHolder(numShards uint32, selfShard uint32) dataRetriever.PoolsHolder {
	var err error

	txPool, err := CreateTxPool(numShards, selfShard)
	panicIfError("CreatePoolsHolder", err)

	unsignedTxPool, err := shardedData.NewShardedData("unsignedTxPool", storageUnit.CacheConfig{
		Capacity:    100000,
		SizeInBytes: 1000000000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolder", err)

	rewardsTxPool, err := shardedData.NewShardedData("rewardsTxPool", storageUnit.CacheConfig{
		Capacity:    300,
		SizeInBytes: 300000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolder", err)

	headersPool, err := headersCache.NewHeadersPool(config.HeadersPoolConfig{
		MaxHeadersPerShard:            1000,
		NumElementsToRemoveOnEviction: 100,
	})
	panicIfError("CreatePoolsHolder", err)

	cacherConfig := storageUnit.CacheConfig{Capacity: 100000, Type: storageUnit.LRUCache, Shards: 1}
	txBlockBody, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	cacherConfig = storageUnit.CacheConfig{Capacity: 100000, Type: storageUnit.LRUCache, Shards: 1}
	peerChangeBlockBody, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	cacherConfig = storageUnit.CacheConfig{Capacity: 50000, Type: storageUnit.LRUCache}
	trieNodes, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolder", err)

	currentTx, err := dataPool.NewCurrentBlockPool()
	panicIfError("CreatePoolsHolder", err)

	holder, err := dataPool.NewDataPool(
		txPool,
		unsignedTxPool,
		rewardsTxPool,
		headersPool,
		txBlockBody,
		peerChangeBlockBody,
		trieNodes,
		currentTx,
	)
	panicIfError("CreatePoolsHolder", err)

	return holder
}

// CreatePoolsHolderWithTxPool -
func CreatePoolsHolderWithTxPool(txPool dataRetriever.ShardedDataCacherNotifier) dataRetriever.PoolsHolder {
	var err error

	unsignedTxPool, err := shardedData.NewShardedData("unsignedTxPool", storageUnit.CacheConfig{
		Capacity:    100000,
		SizeInBytes: 1000000000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolderWithTxPool", err)

	rewardsTxPool, err := shardedData.NewShardedData("rewardsTxPool", storageUnit.CacheConfig{
		Capacity:    300,
		SizeInBytes: 300000,
		Shards:      1,
	})
	panicIfError("CreatePoolsHolderWithTxPool", err)

	headersPool, err := headersCache.NewHeadersPool(config.HeadersPoolConfig{
		MaxHeadersPerShard:            1000,
		NumElementsToRemoveOnEviction: 100,
	})
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig := storageUnit.CacheConfig{Capacity: 100000, Type: storageUnit.LRUCache, Shards: 1}
	txBlockBody, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig = storageUnit.CacheConfig{Capacity: 100000, Type: storageUnit.LRUCache, Shards: 1}
	peerChangeBlockBody, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	cacherConfig = storageUnit.CacheConfig{Capacity: 50000, Type: storageUnit.LRUCache}
	trieNodes, err := storageUnit.NewCache(cacherConfig)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	currentTx, err := dataPool.NewCurrentBlockPool()
	panicIfError("CreatePoolsHolderWithTxPool", err)

	holder, err := dataPool.NewDataPool(
		txPool,
		unsignedTxPool,
		rewardsTxPool,
		headersPool,
		txBlockBody,
		peerChangeBlockBody,
		trieNodes,
		currentTx,
	)
	panicIfError("CreatePoolsHolderWithTxPool", err)

	return holder
}
