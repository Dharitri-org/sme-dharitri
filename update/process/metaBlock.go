package process

import (
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/update"
)

// ArgsNewMetaBlockCreatorAfterHardfork defines the arguments structure for new metablock creator after hardfork
type ArgsNewMetaBlockCreatorAfterHardfork struct {
	ImportHandler     update.ImportHandler
	Marshalizer       marshal.Marshalizer
	Hasher            hashing.Hasher
	ShardCoordinator  sharding.Coordinator
	ValidatorAccounts state.AccountsAdapter
}

type metaBlockCreator struct {
	importHandler     update.ImportHandler
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	validatorAccounts state.AccountsAdapter
}

// NewMetaBlockCreatorAfterHardfork creates the after hardfork metablock creator
func NewMetaBlockCreatorAfterHardfork(args ArgsNewMetaBlockCreatorAfterHardfork) (*metaBlockCreator, error) {
	if check.IfNil(args.ImportHandler) {
		return nil, update.ErrNilImportHandler
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, update.ErrNilHasher
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.ValidatorAccounts) {
		return nil, update.ErrNilAccounts
	}

	return &metaBlockCreator{
		importHandler:     args.ImportHandler,
		marshalizer:       args.Marshalizer,
		hasher:            args.Hasher,
		shardCoordinator:  args.ShardCoordinator,
		validatorAccounts: args.ValidatorAccounts,
	}, nil
}

// CreateNewBlock will create a new block after hardfork import
func (m *metaBlockCreator) CreateNewBlock(
	chainID string,
	round uint64,
	nonce uint64,
	epoch uint32,
) (data.HeaderHandler, data.BodyHandler, error) {
	if len(chainID) == 0 {
		return nil, nil, update.ErrEmptyChainID
	}

	validatorRootHash, err := m.validatorAccounts.Commit()
	if err != nil {
		return nil, nil, err
	}

	accounts := m.importHandler.GetAccountsDBForShard(core.MetachainShardId)
	if check.IfNil(accounts) {
		return nil, nil, update.ErrNilAccounts
	}

	rootHash, err := accounts.Commit()
	if err != nil {
		return nil, nil, err
	}

	hardForkMeta := m.importHandler.GetHardForkMetaBlock()
	blockBody := &block.Body{
		MiniBlocks: make([]*block.MiniBlock, 0),
	}
	metaHdr := &block.MetaBlock{
		Nonce:                  nonce,
		Round:                  round,
		PrevRandSeed:           rootHash,
		RandSeed:               rootHash,
		RootHash:               rootHash,
		ValidatorStatsRootHash: validatorRootHash,
		EpochStart:             hardForkMeta.EpochStart,
		ChainID:                []byte(chainID),
		SoftwareVersion:        []byte(""),
		AccumulatedFees:        big.NewInt(0),
		AccumulatedFeesInEpoch: big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		Epoch:                  epoch,
		PubKeysBitmap:          []byte{1},
	}

	return metaHdr, blockBody, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (m *metaBlockCreator) IsInterfaceNil() bool {
	return m == nil
}
