package factory

import (
	"fmt"
	"path/filepath"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/storage"
	"github.com/Dharitri-org/sme-dharitri/storage/pruning"
	"github.com/Dharitri-org/sme-dharitri/storage/storageUnit"
	logger "github.com/Dharitri-org/sme-logger"
)

var log = logger.GetOrCreate("storage/factory")

const (
	minimumNumberOfActivePersisters = 1
	minimumNumberOfEpochsToKeep     = 2
)

// StorageServiceFactory handles the creation of storage services for both meta and shards
type StorageServiceFactory struct {
	generalConfig      *config.Config
	shardCoordinator   storage.ShardCoordinator
	pathManager        storage.PathManagerHandler
	epochStartNotifier storage.EpochStartNotifier
	currentEpoch       uint32
}

// NewStorageServiceFactory will return a new instance of StorageServiceFactory
func NewStorageServiceFactory(
	config *config.Config,
	shardCoordinator storage.ShardCoordinator,
	pathManager storage.PathManagerHandler,
	epochStartNotifier storage.EpochStartNotifier,
	currentEpoch uint32,
) (*StorageServiceFactory, error) {
	if config == nil {
		return nil, storage.ErrNilConfig
	}
	if config.StoragePruning.NumEpochsToKeep < minimumNumberOfEpochsToKeep && config.StoragePruning.CleanOldEpochsData {
		return nil, storage.ErrInvalidNumberOfEpochsToSave
	}
	if config.StoragePruning.NumActivePersisters < minimumNumberOfActivePersisters {
		return nil, storage.ErrInvalidNumberOfActivePersisters
	}
	if check.IfNil(shardCoordinator) {
		return nil, storage.ErrNilShardCoordinator
	}
	if check.IfNil(pathManager) {
		return nil, storage.ErrNilPathManager
	}
	if check.IfNil(epochStartNotifier) {
		return nil, storage.ErrNilEpochStartNotifier
	}

	return &StorageServiceFactory{
		generalConfig:      config,
		shardCoordinator:   shardCoordinator,
		pathManager:        pathManager,
		epochStartNotifier: epochStartNotifier,
		currentEpoch:       currentEpoch,
	}, nil
}

// CreateForShard will return the storage service which contains all storers needed for a shard
func (psf *StorageServiceFactory) CreateForShard() (dataRetriever.StorageService, error) {
	var headerUnit *pruning.PruningStorer
	var peerBlockUnit *pruning.PruningStorer
	var miniBlockUnit *pruning.PruningStorer
	var txUnit *pruning.PruningStorer
	var metachainHeaderUnit *pruning.PruningStorer
	var unsignedTxUnit *pruning.PruningStorer
	var rewardTxUnit *pruning.PruningStorer
	var bootstrapUnit *pruning.PruningStorer
	var txLogsUnit *pruning.PruningStorer
	var err error

	successfullyCreatedStorers := make([]storage.Storer, 0)
	defer func() {
		// cleanup
		if err != nil {
			for _, storer := range successfullyCreatedStorers {
				_ = storer.DestroyUnit()
			}
		}
	}()

	txUnitStorerArgs := psf.createPruningStorerArgs(psf.generalConfig.TxStorage)
	txUnit, err = pruning.NewPruningStorer(txUnitStorerArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txUnit)

	unsignedTxUnitStorerArgs := psf.createPruningStorerArgs(psf.generalConfig.UnsignedTransactionStorage)
	unsignedTxUnit, err = pruning.NewPruningStorer(unsignedTxUnitStorerArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, unsignedTxUnit)

	rewardTxUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.RewardTxStorage)
	rewardTxUnit, err = pruning.NewPruningStorer(rewardTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, rewardTxUnit)

	miniBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MiniBlocksStorage)
	miniBlockUnit, err = pruning.NewPruningStorer(miniBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockUnit)

	peerBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.PeerBlockBodyStorage)
	peerBlockUnit, err = pruning.NewPruningStorer(peerBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, peerBlockUnit)

	headerUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BlockHeaderStorage)
	headerUnit, err = pruning.NewPruningStorer(headerUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, headerUnit)

	metaChainHeaderUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MetaBlockStorage)
	metachainHeaderUnit, err = pruning.NewPruningStorer(metaChainHeaderUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metachainHeaderUnit)

	// metaHdrHashNonce is static
	metaHdrHashNonceUnitConfig := GetDBFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.DB)
	shardID := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath := psf.pathManager.PathForStatic(shardID, psf.generalConfig.MetaHdrNonceHashStorage.DB.FilePath)
	metaHdrHashNonceUnitConfig.FilePath = dbPath
	metaHdrHashNonceUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Cache),
		metaHdrHashNonceUnitConfig,
		GetBloomFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaHdrHashNonceUnit)

	// shardHdrHashNonce storer is static
	shardHdrHashNonceConfig := GetDBFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.DB)
	shardID = core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath = psf.pathManager.PathForStatic(shardID, psf.generalConfig.ShardHdrNonceHashStorage.DB.FilePath) + shardID
	shardHdrHashNonceConfig.FilePath = dbPath
	shardHdrHashNonceUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Cache),
		shardHdrHashNonceConfig,
		GetBloomFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, shardHdrHashNonceUnit)

	heartbeatDbConfig := GetDBFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.DB)
	shardId := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath = psf.pathManager.PathForStatic(shardId, psf.generalConfig.Heartbeat.HeartbeatStorage.DB.FilePath)
	heartbeatDbConfig.FilePath = dbPath
	heartbeatStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Cache),
		heartbeatDbConfig,
		GetBloomFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, heartbeatStorageUnit)

	statusMetricsDbConfig := GetDBFromConfig(psf.generalConfig.StatusMetricsStorage.DB)
	shardId = core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath = psf.pathManager.PathForStatic(shardId, psf.generalConfig.StatusMetricsStorage.DB.FilePath)
	statusMetricsDbConfig.FilePath = dbPath
	statusMetricsStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.StatusMetricsStorage.Cache),
		statusMetricsDbConfig,
		GetBloomFromConfig(psf.generalConfig.StatusMetricsStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, statusMetricsStorageUnit)

	bootstrapUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BootstrapStorage)
	bootstrapUnit, err = pruning.NewPruningStorer(bootstrapUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, bootstrapUnit)

	txLogsUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.TxLogsStorage)
	txLogsUnit, err = pruning.NewPruningStorer(txLogsUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txLogsUnit)

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, txUnit)
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)
	store.AddStorer(dataRetriever.PeerChangesUnit, peerBlockUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaBlockUnit, metachainHeaderUnit)
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, unsignedTxUnit)
	store.AddStorer(dataRetriever.RewardTransactionUnit, rewardTxUnit)
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(psf.shardCoordinator.SelfId())
	store.AddStorer(hdrNonceHashDataUnit, shardHdrHashNonceUnit)
	store.AddStorer(dataRetriever.HeartbeatUnit, heartbeatStorageUnit)
	store.AddStorer(dataRetriever.BootstrapUnit, bootstrapUnit)
	store.AddStorer(dataRetriever.StatusMetricsUnit, statusMetricsStorageUnit)
	store.AddStorer(dataRetriever.TxLogsUnit, txLogsUnit)

	historyTxUnit, hashEpochUnit, err := psf.createHistoryStorersIfNeeded()
	if err != nil {
		return nil, err
	}

	if psf.generalConfig.FullHistory.Enabled {
		successfullyCreatedStorers = append(successfullyCreatedStorers, historyTxUnit)
		store.AddStorer(dataRetriever.TransactionHistoryUnit, historyTxUnit)

		successfullyCreatedStorers = append(successfullyCreatedStorers, hashEpochUnit)
		store.AddStorer(dataRetriever.EpochByHashUnit, hashEpochUnit)
	}

	return store, err
}

// CreateForMeta will return the storage service which contains all storers needed for metachain
func (psf *StorageServiceFactory) CreateForMeta() (dataRetriever.StorageService, error) {
	var metaBlockUnit *pruning.PruningStorer
	var headerUnit *pruning.PruningStorer
	var txUnit *pruning.PruningStorer
	var miniBlockUnit *pruning.PruningStorer
	var unsignedTxUnit *pruning.PruningStorer
	var rewardTxUnit *pruning.PruningStorer
	var bootstrapUnit *pruning.PruningStorer
	var txLogsUnit *pruning.PruningStorer
	var err error

	successfullyCreatedStorers := make([]storage.Storer, 0)

	defer func() {
		// cleanup
		if err != nil {
			log.Error("create meta store", "error", err.Error())
			for _, storer := range successfullyCreatedStorers {
				_ = storer.DestroyUnit()
			}
		}
	}()

	metaBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MetaBlockStorage)
	metaBlockUnit, err = pruning.NewPruningStorer(metaBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaBlockUnit)

	headerUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BlockHeaderStorage)
	headerUnit, err = pruning.NewPruningStorer(headerUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, headerUnit)

	// metaHdrHashNonce is static
	metaHdrHashNonceUnitConfig := GetDBFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.DB)
	shardID := core.GetShardIDString(core.MetachainShardId)
	dbPath := psf.pathManager.PathForStatic(shardID, psf.generalConfig.MetaHdrNonceHashStorage.DB.FilePath)
	metaHdrHashNonceUnitConfig.FilePath = dbPath
	metaHdrHashNonceUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Cache),
		metaHdrHashNonceUnitConfig,
		GetBloomFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaHdrHashNonceUnit)

	shardHdrHashNonceUnits := make([]*storageUnit.Unit, psf.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
		shardHdrHashNonceConfig := GetDBFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.DB)
		shardID = core.GetShardIDString(core.MetachainShardId)
		dbPath = psf.pathManager.PathForStatic(shardID, psf.generalConfig.ShardHdrNonceHashStorage.DB.FilePath) + fmt.Sprintf("%d", i)
		shardHdrHashNonceConfig.FilePath = dbPath
		shardHdrHashNonceUnits[i], err = storageUnit.NewStorageUnitFromConf(
			GetCacherFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Cache),
			shardHdrHashNonceConfig,
			GetBloomFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Bloom))
		if err != nil {
			return nil, err
		}

		successfullyCreatedStorers = append(successfullyCreatedStorers, shardHdrHashNonceUnits[i])
	}

	shardId := core.GetShardIDString(psf.shardCoordinator.SelfId())
	heartbeatDbConfig := GetDBFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.DB)
	dbPath = psf.pathManager.PathForStatic(shardId, psf.generalConfig.Heartbeat.HeartbeatStorage.DB.FilePath)
	heartbeatDbConfig.FilePath = dbPath
	heartbeatStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Cache),
		heartbeatDbConfig,
		GetBloomFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, heartbeatStorageUnit)

	statusMetricsDbConfig := GetDBFromConfig(psf.generalConfig.StatusMetricsStorage.DB)
	shardId = core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath = psf.pathManager.PathForStatic(shardId, psf.generalConfig.StatusMetricsStorage.DB.FilePath)
	statusMetricsDbConfig.FilePath = dbPath
	statusMetricsStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.StatusMetricsStorage.Cache),
		statusMetricsDbConfig,
		GetBloomFromConfig(psf.generalConfig.StatusMetricsStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, statusMetricsStorageUnit)

	txUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.TxStorage)
	txUnit, err = pruning.NewPruningStorer(txUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txUnit)

	unsignedTxUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.UnsignedTransactionStorage)
	unsignedTxUnit, err = pruning.NewPruningStorer(unsignedTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, unsignedTxUnit)

	rewardTxUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.RewardTxStorage)
	rewardTxUnit, err = pruning.NewPruningStorer(rewardTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, rewardTxUnit)

	miniBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MiniBlocksStorage)
	miniBlockUnit, err = pruning.NewPruningStorer(miniBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockUnit)

	bootstrapUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BootstrapStorage)
	bootstrapUnit, err = pruning.NewPruningStorer(bootstrapUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, bootstrapUnit)

	txLogsUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.TxLogsStorage)
	txLogsUnit, err = pruning.NewPruningStorer(txLogsUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txLogsUnit)

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, metaBlockUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)
	store.AddStorer(dataRetriever.TransactionUnit, txUnit)
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, unsignedTxUnit)
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)
	store.AddStorer(dataRetriever.RewardTransactionUnit, rewardTxUnit)
	for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, shardHdrHashNonceUnits[i])
	}
	store.AddStorer(dataRetriever.HeartbeatUnit, heartbeatStorageUnit)
	store.AddStorer(dataRetriever.BootstrapUnit, bootstrapUnit)
	store.AddStorer(dataRetriever.StatusMetricsUnit, statusMetricsStorageUnit)
	store.AddStorer(dataRetriever.TxLogsUnit, txLogsUnit)

	historyTxUnit, hashEpochUnit, err := psf.createHistoryStorersIfNeeded()
	if err != nil {
		return nil, err
	}

	if psf.generalConfig.FullHistory.Enabled {
		successfullyCreatedStorers = append(successfullyCreatedStorers, historyTxUnit)
		store.AddStorer(dataRetriever.TransactionHistoryUnit, historyTxUnit)

		successfullyCreatedStorers = append(successfullyCreatedStorers, hashEpochUnit)
		store.AddStorer(dataRetriever.EpochByHashUnit, hashEpochUnit)
	}

	return store, err
}

func (psf *StorageServiceFactory) createHistoryStorersIfNeeded() (*pruning.PruningStorer, *storageUnit.Unit, error) {
	if !psf.generalConfig.FullHistory.Enabled {
		return nil, nil, nil
	}

	historyTxsUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.FullHistory.HistoryTransactionStorageConfig)
	historyTxUnit, err := pruning.NewPruningStorer(historyTxsUnitArgs)
	if err != nil {
		return nil, nil, err
	}

	hashEpochDbConfig := GetDBFromConfig(psf.generalConfig.FullHistory.HashEpochStorageConfig.DB)
	shardID := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath := psf.pathManager.PathForStatic(shardID, psf.generalConfig.FullHistory.HashEpochStorageConfig.DB.FilePath)
	hashEpochDbConfig.FilePath = dbPath
	hashEpochUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.FullHistory.HashEpochStorageConfig.Cache),
		hashEpochDbConfig,
		GetBloomFromConfig(psf.generalConfig.FullHistory.HashEpochStorageConfig.Bloom))
	if err != nil {
		return nil, nil, err
	}

	return historyTxUnit, hashEpochUnit, nil
}

func (psf *StorageServiceFactory) createPruningStorerArgs(storageConfig config.StorageConfig) *pruning.StorerArgs {
	cleanOldEpochsData := psf.generalConfig.StoragePruning.CleanOldEpochsData
	numOfEpochsToKeep := uint32(psf.generalConfig.StoragePruning.NumEpochsToKeep)
	numOfActivePersisters := uint32(psf.generalConfig.StoragePruning.NumActivePersisters)
	pruningEnabled := psf.generalConfig.StoragePruning.Enabled
	shardId := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath := filepath.Join(psf.pathManager.PathForEpoch(shardId, psf.currentEpoch, storageConfig.DB.FilePath))
	args := &pruning.StorerArgs{
		Identifier:            storageConfig.DB.FilePath,
		PruningEnabled:        pruningEnabled,
		StartingEpoch:         psf.currentEpoch,
		CleanOldEpochsData:    cleanOldEpochsData,
		ShardCoordinator:      psf.shardCoordinator,
		CacheConf:             GetCacherFromConfig(storageConfig.Cache),
		PathManager:           psf.pathManager,
		DbPath:                dbPath,
		PersisterFactory:      NewPersisterFactory(storageConfig.DB),
		BloomFilterConf:       GetBloomFromConfig(storageConfig.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
		MaxBatchSize:          storageConfig.DB.MaxBatchSize,
	}

	return args
}
