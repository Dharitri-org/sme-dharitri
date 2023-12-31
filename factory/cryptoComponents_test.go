package factory_test

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/factory"
	"github.com/Dharitri-org/sme-dharitri/factory/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCryptoComponentsFactory_NiNodesConfigShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.NodesConfig = nil
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilNodesConfig, err)
}

func TestNewCryptoComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.ShardCoordinator = nil
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilShardCoordinator, err)
}

func TestNewCryptoComponentsFactory_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.KeyGen = nil
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilKeyGen, err)
}

func TestNewCryptoComponentsFactory_NilPrivateKeyShouldErr(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.PrivKey = nil
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.Nil(t, ccf)
	require.Equal(t, factory.ErrNilPrivateKey, err)
}

func TestNewCryptoComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	ccf, err := factory.NewCryptoComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, ccf)
}

func TestCryptoComponentsFactory_CreateShouldErrDueToBadConfig(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	args.Config = config.Config{}
	ccf, _ := factory.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.Error(t, err)
	require.Nil(t, cc)
}

func TestCryptoComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	args := getCryptoArgs()
	ccf, _ := factory.NewCryptoComponentsFactory(args)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func getCryptoArgs() factory.CryptoComponentsFactoryArgs {
	return factory.CryptoComponentsFactoryArgs{
		Config: config.Config{
			Hasher:         config.TypeConfig{Type: "blake2b"},
			MultisigHasher: config.TypeConfig{Type: "blake2b"},
			PublicKeyPIDSignature: config.CacheConfig{
				Capacity: 1000,
				Type:     "LRU",
			},
			Consensus: config.TypeConfig{Type: "bls"},
		},
		NodesConfig:      &mock.NodesSetupStub{},
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(2),
		KeyGen:           &mock.KeyGenMock{},
		PrivKey:          &mock.PrivateKeyMock{},
	}
}
