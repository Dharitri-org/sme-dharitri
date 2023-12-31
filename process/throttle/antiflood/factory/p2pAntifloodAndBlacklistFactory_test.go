package factory

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/p2p"
	"github.com/Dharitri-org/sme-dharitri/p2p/mock"
	"github.com/Dharitri-org/sme-dharitri/process/throttle/antiflood/disabled"
	"github.com/stretchr/testify/assert"
)

const currentPid = core.PeerID("current pid")

func TestNewP2PAntiFloodAndBlackList_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	af, pids, pks, err := NewP2PAntiFloodAndBlackList(cfg, nil, currentPid)
	assert.Nil(t, af)
	assert.Nil(t, pids)
	assert.Nil(t, pks)
	assert.Equal(t, p2p.ErrNilStatusHandler, err)
}

func TestNewP2PAntiFloodAndBlackList_ShouldWorkAndReturnDisabledImplementations(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Antiflood: config.AntifloodConfig{
			Enabled: false,
		},
	}
	ash := &mock.AppStatusHandlerMock{}
	af, pids, pks, err := NewP2PAntiFloodAndBlackList(cfg, ash, currentPid)
	assert.NotNil(t, af)
	assert.NotNil(t, pids)
	assert.NotNil(t, pks)
	assert.Nil(t, err)

	_, ok1 := af.(*disabled.AntiFlood)
	_, ok2 := pids.(*disabled.PeerBlacklistCacher)
	_, ok3 := pks.(*disabled.TimeCache)
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.True(t, ok3)
}

func TestNewP2PAntiFloodAndBlackList_ShouldWorkAndReturnOkImplementations(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Antiflood: config.AntifloodConfig{
			Enabled: true,
			Cache: config.CacheConfig{
				Type:     "LRU",
				Capacity: 10,
				Shards:   2,
			},
			FastReacting: createFloodPreventerConfig(),
			SlowReacting: createFloodPreventerConfig(),
			OutOfSpecs:   createFloodPreventerConfig(),
			Topic: config.TopicAntifloodConfig{
				DefaultMaxMessagesPerSec: 10,
			},
		},
	}

	ash := &mock.AppStatusHandlerMock{}
	af, pids, pks, err := NewP2PAntiFloodAndBlackList(cfg, ash, currentPid)
	assert.Nil(t, err)
	assert.NotNil(t, af)
	assert.NotNil(t, pids)
	assert.NotNil(t, pks)
}

func createFloodPreventerConfig() config.FloodPreventerConfig {
	return config.FloodPreventerConfig{
		IntervalInSeconds: 1,
		PeerMaxInput: config.AntifloodLimitsConfig{
			BaseMessagesPerInterval: 10,
			TotalSizePerInterval:    10,
		},
		BlackList: config.BlackListConfig{
			ThresholdNumMessagesPerInterval: 10,
			ThresholdSizePerInterval:        10,
			NumFloodingRounds:               10,
			PeerBanDurationInSeconds:        10,
		},
	}
}
