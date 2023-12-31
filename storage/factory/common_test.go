package factory

import (
	"fmt"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func TestGetCacherFromConfig(t *testing.T) {
	t.Parallel()

	cfg := config.CacheConfig{
		Capacity:    100,
		Shards:      2,
		Type:        "lru",
		SizeInBytes: 128,
	}

	storageCacheConfig := GetCacherFromConfig(cfg)
	assert.Equal(t, storageUnit.CacheConfig{
		Capacity:    cfg.Capacity,
		SizeInBytes: cfg.SizeInBytes,
		Type:        storageUnit.CacheType(cfg.Type),
		Shards:      cfg.Shards,
	}, storageCacheConfig)
}

func TestGetDBFromConfig(t *testing.T) {
	t.Parallel()

	cfg := config.DBConfig{
		Type:              "lru",
		MaxBatchSize:      10,
		BatchDelaySeconds: 2,
		MaxOpenFiles:      20,
	}

	storageDBConfig := GetDBFromConfig(cfg)
	assert.Equal(t, storageUnit.DBConfig{
		Type:              storageUnit.DBType(cfg.Type),
		MaxBatchSize:      cfg.MaxBatchSize,
		BatchDelaySeconds: cfg.BatchDelaySeconds,
		MaxOpenFiles:      cfg.MaxOpenFiles,
	}, storageDBConfig)
}

func TestGetBloomFromConfig(t *testing.T) {
	t.Parallel()

	cfg := config.BloomFilterConfig{
		Size:     100,
		HashFunc: []string{"hashFunc"},
	}

	storageBloomConfig := GetBloomFromConfig(cfg)
	assert.Equal(t, storageUnit.BloomConfig{
		HashFunc: []storageUnit.HasherType{storageUnit.HasherType(cfg.HashFunc[0])},
		Size:     cfg.Size,
	}, storageBloomConfig)
}

func TestConvertShardIDToUint32(t *testing.T) {
	t.Parallel()

	shardID, err := convertShardIDToUint32("metachain")
	assert.NoError(t, err)
	assert.Equal(t, core.MetachainShardId, shardID)

	id := uint32(0)
	shardIDStr := fmt.Sprintf("%d", id)
	shardID, err = convertShardIDToUint32(shardIDStr)
	assert.NoError(t, err)
	assert.Equal(t, id, shardID)

	shardID, err = convertShardIDToUint32("wrongID")
	assert.Error(t, err)
	assert.Equal(t, uint32(0), shardID)
}
