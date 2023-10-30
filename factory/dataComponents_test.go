package factory_test

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/factory"
	"github.com/Dharitri-org/sme-dharitri/factory/mock"
	"github.com/Dharitri-org/sme-dharitri/process/economics"
	"github.com/Dharitri-org/sme-dharitri/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewDataComponentsFactory_NilEconomicsDataShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.EconomicsData = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, factory.ErrNilEconomicsData, err)
}

func TestNewDataComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.ShardCoordinator = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, factory.ErrNilShardCoordinator, err)
}

func TestNewDataComponentsFactory_NilCoreComponentsShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.Core = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, factory.ErrNilCoreComponents, err)
}

func TestNewDataComponentsFactory_NilPathManagerShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.PathManager = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, factory.ErrNilPathManager, err)
}

func TestNewDataComponentsFactory_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.EpochStartNotifier = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, factory.ErrNilEpochStartNotifier, err)
}

func TestNewDataComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getDataArgs()

	dcf, err := factory.NewDataComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, dcf)
}

func TestDataComponentsFactory_CreateShouldErrDueBadConfig(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
	dcf, err := factory.NewDataComponentsFactory(args)
	require.NoError(t, err)

	dc, err := dcf.Create()
	require.Error(t, err)
	require.Nil(t, dc)
}

func TestDataComponentsFactory_CreateForShardShouldWork(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	dcf, err := factory.NewDataComponentsFactory(args)

	require.NoError(t, err)
	dc, err := dcf.Create()
	require.NoError(t, err)
	require.NotNil(t, dc)
}

func TestDataComponentsFactory_CreateForMetaShouldWork(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	multiShrdCoord := mock.NewMultiShardsCoordinatorMock(3)
	multiShrdCoord.CurrentShard = core.MetachainShardId
	args.ShardCoordinator = multiShrdCoord
	dcf, err := factory.NewDataComponentsFactory(args)
	require.NoError(t, err)
	dc, err := dcf.Create()
	require.NoError(t, err)
	require.NotNil(t, dc)
}

func getDataArgs() factory.DataComponentsFactoryArgs {
	testEconomics := &economics.TestEconomicsData{EconomicsData: &economics.EconomicsData{}}
	testEconomics.SetMinGasPrice(200000000000)

	return factory.DataComponentsFactoryArgs{
		Config:             testscommon.GetGeneralConfig(),
		EconomicsData:      testEconomics.EconomicsData,
		ShardCoordinator:   mock.NewMultiShardsCoordinatorMock(2),
		Core:               getCoreComponents(),
		PathManager:        &mock.PathManagerStub{},
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		CurrentEpoch:       0,
	}
}
