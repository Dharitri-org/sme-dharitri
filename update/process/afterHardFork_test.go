package process

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/update"
	"github.com/Dharitri-org/sme-dharitri/update/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArgsAfterHardFork() ArgsAfterHardFork {
	return ArgsAfterHardFork{
		MapBlockProcessors: make(map[uint32]update.HardForkBlockProcessor),
		ImportHandler:      &mock.ImportHandlerStub{},
		ShardCoordinator:   mock.NewOneShardCoordinatorMock(),
		Hasher:             &mock.HasherMock{},
		Marshalizer:        &mock.MarshalizerMock{},
	}
}

func TestNewAfterHardForkBlockCreation(t *testing.T) {
	t.Parallel()

	args := createMockArgsAfterHardFork()

	hardForkBlockCreator, err := NewAfterHardForkBlockCreation(args)
	assert.NoError(t, err)
	assert.False(t, check.IfNil(hardForkBlockCreator))
}

func TestCreateAllBlocksAfterHardfork(t *testing.T) {
	t.Parallel()

	args := createMockArgsAfterHardFork()
	args.ShardCoordinator = &mock.CoordinatorStub{
		NumberOfShardsCalled: func() uint32 {
			return 1
		},
	}

	hdr1 := &block.Header{}
	hdr2 := &block.Header{}
	body1 := &block.Body{}
	body2 := &block.Body{}
	args.MapBlockProcessors[0] = &mock.HardForkBlockProcessor{
		CreateNewBlockCalled: func(chainID string, round uint64, nonce uint64, epoch uint32) (data.HeaderHandler, data.BodyHandler, error) {
			return hdr1, body1, nil
		},
	}
	args.MapBlockProcessors[core.MetachainShardId] = &mock.HardForkBlockProcessor{
		CreateNewBlockCalled: func(chainID string, round uint64, nonce uint64, epoch uint32) (data.HeaderHandler, data.BodyHandler, error) {
			return hdr2, body2, nil
		},
	}

	hardForkBlockCreator, _ := NewAfterHardForkBlockCreation(args)

	expectedHeaders := map[uint32]data.HeaderHandler{
		0: hdr1, core.MetachainShardId: hdr2,
	}
	expectedBodies := map[uint32]data.BodyHandler{
		0: body1, core.MetachainShardId: body2,
	}
	chainID, round, nonce, epoch := "chainId", uint64(100), uint64(90), uint32(2)
	headers, bodies, err := hardForkBlockCreator.CreateAllBlocksAfterHardfork(chainID, round, nonce, epoch)
	assert.NoError(t, err)
	assert.Equal(t, expectedHeaders, headers)
	assert.Equal(t, expectedBodies, bodies)

}
