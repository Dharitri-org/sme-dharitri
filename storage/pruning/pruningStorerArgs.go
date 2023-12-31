package pruning

import (
	"github.com/Dharitri-org/sme-dharitri/storage"
	"github.com/Dharitri-org/sme-dharitri/storage/storageUnit"
)

// StorerArgs will hold the arguments needed for PruningStorer
type StorerArgs struct {
	Identifier            string
	ShardCoordinator      storage.ShardCoordinator
	CacheConf             storageUnit.CacheConfig
	PathManager           storage.PathManagerHandler
	DbPath                string
	PersisterFactory      DbFactoryHandler
	BloomFilterConf       storageUnit.BloomConfig
	Notifier              EpochStartNotifier
	NumOfEpochsToKeep     uint32
	NumOfActivePersisters uint32
	StartingEpoch         uint32
	MaxBatchSize          int
	PruningEnabled        bool
	CleanOldEpochsData    bool
}
