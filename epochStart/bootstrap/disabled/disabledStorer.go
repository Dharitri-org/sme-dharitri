package disabled

import (
	"github.com/Dharitri-org/sme-dharitri/storage"
	"github.com/Dharitri-org/sme-dharitri/storage/memorydb"
	"github.com/Dharitri-org/sme-dharitri/storage/storageUnit"
)

const defaultCapacity = 10000
const defaultNumShards = 1
const zeroSize = 0

// CreateMemUnit creates an in-memory storer unit using maps
func CreateMemUnit() storage.Storer {
	cache, err := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: defaultCapacity, Shards: defaultNumShards, SizeInBytes: zeroSize})
	if err != nil {
		return nil
	}

	unit, err := storageUnit.NewStorageUnit(cache, memorydb.New())
	if err != nil {
		return nil
	}

	return unit
}
