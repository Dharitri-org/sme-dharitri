package factory

import (
	"errors"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/storage"
	"github.com/Dharitri-org/sme-dharitri/storage/leveldb"
	"github.com/Dharitri-org/sme-dharitri/storage/memorydb"
	"github.com/Dharitri-org/sme-dharitri/storage/storageUnit"
)

// PersisterFactory is the factory which will handle creating new databases
type PersisterFactory struct {
	dbType            string
	batchDelaySeconds int
	maxBatchSize      int
	maxOpenFiles      int
}

// NewPersisterFactory will return a new instance of a PersisterFactory
func NewPersisterFactory(config config.DBConfig) *PersisterFactory {
	return &PersisterFactory{
		dbType:            config.Type,
		batchDelaySeconds: config.BatchDelaySeconds,
		maxBatchSize:      config.MaxBatchSize,
		maxOpenFiles:      config.MaxOpenFiles,
	}
}

// Create will return a new instance of a DB with a given path
func (pf *PersisterFactory) Create(path string) (storage.Persister, error) {
	if len(path) == 0 {
		return nil, errors.New("invalid file path")
	}

	switch storageUnit.DBType(pf.dbType) {
	case storageUnit.LvlDB:
		return leveldb.NewDB(path, pf.batchDelaySeconds, pf.maxBatchSize, pf.maxOpenFiles)
	case storageUnit.LvlDBSerial:
		return leveldb.NewSerialDB(path, pf.batchDelaySeconds, pf.maxBatchSize, pf.maxOpenFiles)
	case storageUnit.MemoryDB:
		return memorydb.New(), nil
	default:
		return nil, storage.ErrNotSupportedDBType
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (pf *PersisterFactory) IsInterfaceNil() bool {
	return pf == nil
}
