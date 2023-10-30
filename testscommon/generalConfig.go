package testscommon

import (
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/storage/storageUnit"
)

func GetGeneralConfig() config.Config {
	return config.Config{
		GeneralSettings: config.GeneralSettingsConfig{
			StartInEpochEnabled: true,
		},
		EpochStartConfig: config.EpochStartConfig{
			MinRoundsBetweenEpochs:            5,
			RoundsPerEpoch:                    10,
			MinNumConnectedPeersToStart:       2,
			MinNumOfPeersToConsiderBlockValid: 2,
		},
		WhiteListPool:          getLRUCacheConfig(),
		WhiteListerVerifiedTxs: getLRUCacheConfig(),
		StoragePruning: config.StoragePruningConfig{
			Enabled:             false,
			CleanOldEpochsData:  false,
			NumEpochsToKeep:     3,
			NumActivePersisters: 3,
		},
		EvictionWaitingList: config.EvictionWaitingListConfig{
			Size: 100,
			DB: config.DBConfig{
				FilePath:          "EvictionWaitingList",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		TrieSnapshotDB: config.DBConfig{
			FilePath:          "TrieSnapshot",
			Type:              string(storageUnit.MemoryDB),
			BatchDelaySeconds: 30,
			MaxBatchSize:      6,
			MaxOpenFiles:      10,
		},
		AccountsTrieStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "AccountsTrie/MainDB",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerAccountsTrieStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "PeerAccountsTrie/MainDB",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		StateTriesConfig: config.StateTriesConfig{
			CheckpointRoundsModulus:     100,
			AccountsStatePruningEnabled: false,
			PeerStatePruningEnabled:     false,
			MaxStateTrieLevelInMemory:   5,
			MaxPeerTrieLevelInMemory:    5,
		},
		TrieStorageManagerConfig: config.TrieStorageManagerConfig{
			PruningBufferLen:   1000,
			SnapshotsBufferLen: 10,
			MaxSnapshots:       2,
		},
		TxDataPool: config.CacheConfig{
			Capacity:             10000,
			SizePerSender:        1000,
			SizeInBytes:          1000000000,
			SizeInBytesPerSender: 10000000,
			Shards:               1,
		},
		UnsignedTransactionDataPool: config.CacheConfig{
			Capacity:    10000,
			SizeInBytes: 1000000000,
			Shards:      1,
		},
		RewardTransactionDataPool: config.CacheConfig{
			Capacity:    10000,
			SizeInBytes: 1000000000,
			Shards:      1,
		},
		HeadersPoolConfig: config.HeadersPoolConfig{
			MaxHeadersPerShard:            100,
			NumElementsToRemoveOnEviction: 1,
		},
		TxBlockBodyDataPool:   getLRUCacheConfig(),
		PeerBlockBodyDataPool: getLRUCacheConfig(),
		TrieNodesDataPool:     getLRUCacheConfig(),
		TxStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "Transactions",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MiniBlocksStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "MiniBlocks",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		ShardHdrNonceHashStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "ShardHdrHashNonce",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MetaBlockStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "MetaBlock",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		MetaHdrNonceHashStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "MetaHdrHashNonce",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		UnsignedTransactionStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "UnsignedTransactions",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		RewardTxStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "RewardTransactions",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		BlockHeaderStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "BlockHeaders",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		Heartbeat: config.HeartbeatConfig{
			HeartbeatStorage: config.StorageConfig{
				Cache: getLRUCacheConfig(),
				DB: config.DBConfig{
					FilePath:          "HeartbeatStorage",
					Type:              string(storageUnit.MemoryDB),
					BatchDelaySeconds: 30,
					MaxBatchSize:      6,
					MaxOpenFiles:      10,
				},
			},
		},
		StatusMetricsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "StatusMetricsStorageDB",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		PeerBlockBodyStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "PeerBlocks",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 30,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		BootstrapStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "BootstrapData",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 1,
				MaxBatchSize:      6,
				MaxOpenFiles:      10,
			},
		},
		TxLogsStorage: config.StorageConfig{
			Cache: getLRUCacheConfig(),
			DB: config.DBConfig{
				FilePath:          "Logs",
				Type:              string(storageUnit.MemoryDB),
				BatchDelaySeconds: 2,
				MaxBatchSize:      100,
				MaxOpenFiles:      10,
			},
		},
	}
}

func getLRUCacheConfig() config.CacheConfig {
	return config.CacheConfig{
		Type:     "LRU",
		Capacity: 1000,
		Shards:   1,
	}
}
