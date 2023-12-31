package metachain

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/display"
	"github.com/Dharitri-org/sme-dharitri/epochStart"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

var _ process.EndOfEpochEconomics = (*economics)(nil)

const numberOfDaysInYear = 365.0
const numberOfSecondsInDay = 86400

type economics struct {
	marshalizer        marshal.Marshalizer
	hasher             hashing.Hasher
	store              dataRetriever.StorageService
	shardCoordinator   sharding.Coordinator
	rewardsHandler     process.RewardsHandler
	roundTime          process.RoundTimeDurationHandler
	genesisEpoch       uint32
	genesisNonce       uint64
	genesisTotalSupply *big.Int
}

// ArgsNewEpochEconomics is the argument for the economics constructor
type ArgsNewEpochEconomics struct {
	Marshalizer        marshal.Marshalizer
	Hasher             hashing.Hasher
	Store              dataRetriever.StorageService
	ShardCoordinator   sharding.Coordinator
	RewardsHandler     process.RewardsHandler
	RoundTime          process.RoundTimeDurationHandler
	GenesisEpoch       uint32
	GenesisNonce       uint64
	GenesisTotalSupply *big.Int
}

// NewEndOfEpochEconomicsDataCreator creates a new end of epoch economics data creator object
func NewEndOfEpochEconomicsDataCreator(args ArgsNewEpochEconomics) (*economics, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, epochStart.ErrNilHasher
	}
	if check.IfNil(args.Store) {
		return nil, epochStart.ErrNilStorage
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, epochStart.ErrNilShardCoordinator
	}
	if check.IfNil(args.RewardsHandler) {
		return nil, epochStart.ErrNilRewardsHandler
	}
	if check.IfNil(args.RoundTime) {
		return nil, process.ErrNilRounder
	}
	if args.GenesisTotalSupply == nil {
		return nil, epochStart.ErrNilGenesisTotalSupply
	}

	e := &economics{
		marshalizer:        args.Marshalizer,
		hasher:             args.Hasher,
		store:              args.Store,
		shardCoordinator:   args.ShardCoordinator,
		rewardsHandler:     args.RewardsHandler,
		roundTime:          args.RoundTime,
		genesisEpoch:       args.GenesisEpoch,
		genesisNonce:       args.GenesisNonce,
		genesisTotalSupply: big.NewInt(0).Set(args.GenesisTotalSupply),
	}

	return e, nil
}

// ComputeEndOfEpochEconomics calculates the rewards per block value for the current epoch
func (e *economics) ComputeEndOfEpochEconomics(
	metaBlock *block.MetaBlock,
) (*block.Economics, error) {
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilHeaderHandler
	}
	if metaBlock.AccumulatedFeesInEpoch == nil {
		return nil, epochStart.ErrNilTotalAccumulatedFeesInEpoch
	}
	if metaBlock.DevFeesInEpoch == nil {
		return nil, epochStart.ErrNilTotalDevFeesInEpoch
	}
	if !metaBlock.IsStartOfEpochBlock() || metaBlock.Epoch < e.genesisEpoch+1 {
		return nil, epochStart.ErrNotEpochStartBlock
	}

	noncesPerShardPrevEpoch, prevEpochStart, err := e.startNoncePerShardFromEpochStart(metaBlock.Epoch - 1)
	if err != nil {
		return nil, err
	}
	prevEpochEconomics := prevEpochStart.EpochStart.Economics

	noncesPerShardCurrEpoch, err := e.startNoncePerShardFromLastCrossNotarized(metaBlock.GetNonce(), metaBlock.EpochStart)
	if err != nil {
		return nil, err
	}

	roundsPassedInEpoch := metaBlock.GetRound() - prevEpochStart.GetRound()
	maxBlocksInEpoch := core.MaxUint64(1, roundsPassedInEpoch*uint64(e.shardCoordinator.NumberOfShards()+1))
	totalNumBlocksInEpoch := e.computeNumOfTotalCreatedBlocks(noncesPerShardPrevEpoch, noncesPerShardCurrEpoch)

	inflationRate := e.computeInflationRate(metaBlock.GetRound())
	rwdPerBlock := e.computeRewardsPerBlock(e.genesisTotalSupply, maxBlocksInEpoch, inflationRate)
	totalRewardsToBeDistributed := big.NewInt(0).Mul(rwdPerBlock, big.NewInt(0).SetUint64(totalNumBlocksInEpoch))

	newTokens := big.NewInt(0).Sub(totalRewardsToBeDistributed, metaBlock.AccumulatedFeesInEpoch)
	if newTokens.Cmp(big.NewInt(0)) < 0 {
		newTokens = big.NewInt(0)
		totalRewardsToBeDistributed = big.NewInt(0).Set(metaBlock.AccumulatedFeesInEpoch)
		rwdPerBlock.Div(totalRewardsToBeDistributed, big.NewInt(0).SetUint64(totalNumBlocksInEpoch))
	}

	e.adjustRewardsPerBlockWithDeveloperFees(rwdPerBlock, metaBlock.DevFeesInEpoch, totalNumBlocksInEpoch)
	e.adjustRewardsPerBlockWithLeaderPercentage(rwdPerBlock, metaBlock.AccumulatedFeesInEpoch, totalNumBlocksInEpoch)
	rewardsForProtocolSustainability := e.computeRewardsForProtocolSustainability(totalRewardsToBeDistributed)
	// adjust rewards per block taking into consideration protocol sustainability rewards
	e.adjustRewardsPerBlockWithProtocolSustainabilityRewards(rwdPerBlock, rewardsForProtocolSustainability, totalNumBlocksInEpoch)

	prevEpochStartHash, err := core.CalculateHash(e.marshalizer, e.hasher, prevEpochStart)
	if err != nil {
		return nil, err
	}

	computedEconomics := block.Economics{
		TotalSupply:                      big.NewInt(0).Add(prevEpochEconomics.TotalSupply, newTokens),
		TotalToDistribute:                big.NewInt(0).Set(totalRewardsToBeDistributed),
		TotalNewlyMinted:                 big.NewInt(0).Set(newTokens),
		RewardsPerBlock:                  rwdPerBlock,
		RewardsForProtocolSustainability: rewardsForProtocolSustainability,
		// TODO: get actual nodePrice from auction smart contract (currently on another feature branch, and not all features enabled)
		NodePrice:           big.NewInt(0).Set(prevEpochEconomics.NodePrice),
		PrevEpochStartRound: prevEpochStart.GetRound(),
		PrevEpochStartHash:  prevEpochStartHash,
	}

	e.printEconomicsData(
		metaBlock,
		prevEpochEconomics,
		inflationRate,
		newTokens,
		computedEconomics,
		totalRewardsToBeDistributed,
		totalNumBlocksInEpoch,
		rwdPerBlock,
		rewardsForProtocolSustainability,
	)

	return &computedEconomics, nil
}

func (e *economics) printEconomicsData(
	metaBlock *block.MetaBlock,
	prevEpochEconomics block.Economics,
	inflationRate float64,
	newTokens *big.Int,
	computedEconomics block.Economics,
	totalRewardsToBeDistributed *big.Int,
	totalNumBlocksInEpoch uint64,
	rwdPerBlock *big.Int,
	rewardsForProtocolSustainability *big.Int,
) {
	header := []string{"identifier", "", "value"}

	rewardsForLeaders := core.GetPercentageOfValue(metaBlock.AccumulatedFeesInEpoch, e.rewardsHandler.LeaderPercentage())
	maxSupplyLength := len(prevEpochEconomics.TotalSupply.String())
	lines := []*display.LineData{
		e.newDisplayLine("epoch", "",
			e.alignRight(fmt.Sprintf("%d", metaBlock.Epoch), maxSupplyLength)),
		e.newDisplayLine("inflation rate", "",
			e.alignRight(fmt.Sprintf("%.6f", inflationRate), maxSupplyLength)),
		e.newDisplayLine("previous total supply", "(1)",
			e.alignRight(prevEpochEconomics.TotalSupply.String(), maxSupplyLength)),
		e.newDisplayLine("new tokens", "(2)",
			e.alignRight(newTokens.String(), maxSupplyLength)),
		e.newDisplayLine("current total supply", "(1+2)",
			e.alignRight(computedEconomics.TotalSupply.String(), maxSupplyLength)),
		e.newDisplayLine("accumulated fees in epoch", "(3)",
			e.alignRight(metaBlock.AccumulatedFeesInEpoch.String(), maxSupplyLength)),
		e.newDisplayLine("total rewards to be distributed", "(4)",
			e.alignRight(totalRewardsToBeDistributed.String(), maxSupplyLength)),
		e.newDisplayLine("total num blocks in epoch", "(5)",
			e.alignRight(fmt.Sprintf("%d", totalNumBlocksInEpoch), maxSupplyLength)),
		e.newDisplayLine("dev fees in epoch", "(6)",
			e.alignRight(metaBlock.DevFeesInEpoch.String(), maxSupplyLength)),
		e.newDisplayLine("leader fees in epoch", "(7)",
			e.alignRight(rewardsForLeaders.String(), maxSupplyLength)),
		e.newDisplayLine("reward per block", "(8)",
			e.alignRight(rwdPerBlock.String(), maxSupplyLength)),
		e.newDisplayLine("percent for protocol sustainability", "(9)",
			e.alignRight(fmt.Sprintf("%.6f", e.rewardsHandler.ProtocolSustainabilityPercentage()), maxSupplyLength)),
		e.newDisplayLine("reward for protocol sustainability", "(4 * 9)",
			e.alignRight(rewardsForProtocolSustainability.String(), maxSupplyLength)),
	}

	str, err := display.CreateTableString(header, lines)
	if err != nil {
		log.Error("economics.printEconomicsData", "error", err)
		return
	}

	log.Debug("computed economics data\n" + str)
}

func (e *economics) alignRight(val string, maxLen int) string {
	if len(val) >= maxLen {
		return val
	}

	return strings.Repeat(" ", maxLen-len(val)) + val
}

func (e *economics) newDisplayLine(values ...string) *display.LineData {
	return display.NewLineData(false, values)
}

// compute the rewards for protocol sustainability - percentage from total rewards
func (e *economics) computeRewardsForProtocolSustainability(totalRewards *big.Int) *big.Int {
	rewardsForProtocolSustainability := core.GetPercentageOfValue(totalRewards, e.rewardsHandler.ProtocolSustainabilityPercentage())
	return rewardsForProtocolSustainability
}

// adjustment for rewards given for each proposed block taking protocol sustainability rewards into consideration
func (e *economics) adjustRewardsPerBlockWithProtocolSustainabilityRewards(
	rwdPerBlock *big.Int,
	protocolSustainabilityRewards *big.Int,
	blocksInEpoch uint64,
) {
	protocolSustainabilityRewardsPerBlock := big.NewInt(0).Div(protocolSustainabilityRewards, big.NewInt(0).SetUint64(blocksInEpoch))
	rwdPerBlock.Sub(rwdPerBlock, protocolSustainabilityRewardsPerBlock)
}

// adjustment for rewards given for each proposed block taking developer fees into consideration
func (e *economics) adjustRewardsPerBlockWithDeveloperFees(
	rwdPerBlock *big.Int,
	developerFees *big.Int,
	blocksInEpoch uint64,
) {
	developerFeesPerBlock := big.NewInt(0).Div(developerFees, big.NewInt(0).SetUint64(blocksInEpoch))
	rwdPerBlock.Sub(rwdPerBlock, developerFeesPerBlock)
}

func (e *economics) adjustRewardsPerBlockWithLeaderPercentage(
	rwdPerBlock *big.Int,
	accumulatedFees *big.Int,
	blocksInEpoch uint64,
) {
	rewardsForLeaders := core.GetPercentageOfValue(accumulatedFees, e.rewardsHandler.LeaderPercentage())
	averageLeaderRewardPerBlock := big.NewInt(0).Div(rewardsForLeaders, big.NewInt(0).SetUint64(blocksInEpoch))
	rwdPerBlock.Sub(rwdPerBlock, averageLeaderRewardPerBlock)
}

// compute inflation rate from genesisTotalSupply and economics settings for that year
func (e *economics) computeInflationRate(currentRound uint64) float64 {
	roundsPerDay := numberOfSecondsInDay / uint64(e.roundTime.TimeDuration().Seconds())
	roundsPerYear := numberOfDaysInYear * roundsPerDay
	yearsIndex := uint32(currentRound/roundsPerYear) + 1
	return e.rewardsHandler.MaxInflationRate(yearsIndex)
}

// compute rewards per block from according to inflation rate and total supply from previous block and maxBlocksPerEpoch
func (e *economics) computeRewardsPerBlock(
	prevTotalSupply *big.Int,
	maxBlocksInEpoch uint64,
	inflationRate float64,
) *big.Int {

	inflationRatePerDay := inflationRate / numberOfDaysInYear
	roundsPerDay := numberOfSecondsInDay / uint64(e.roundTime.TimeDuration().Seconds())
	maxBlocksInADay := core.MaxUint64(1, roundsPerDay*uint64(e.shardCoordinator.NumberOfShards()+1))

	inflationRateForEpoch := inflationRatePerDay * (float64(maxBlocksInEpoch) / float64(maxBlocksInADay))

	rewardsPerBlock := big.NewInt(0).Div(prevTotalSupply, big.NewInt(0).SetUint64(maxBlocksInEpoch))
	rewardsPerBlock = core.GetPercentageOfValue(rewardsPerBlock, inflationRateForEpoch)

	return rewardsPerBlock
}

func (e *economics) computeNumOfTotalCreatedBlocks(
	mapStartNonce map[uint32]uint64,
	mapEndNonce map[uint32]uint64,
) uint64 {
	totalNumBlocks := uint64(0)
	for shardId := uint32(0); shardId < e.shardCoordinator.NumberOfShards(); shardId++ {
		totalNumBlocks += mapEndNonce[shardId] - mapStartNonce[shardId]
	}
	totalNumBlocks += mapEndNonce[core.MetachainShardId] - mapStartNonce[core.MetachainShardId]

	return core.MaxUint64(1, totalNumBlocks)
}

func (e *economics) startNoncePerShardFromEpochStart(epoch uint32) (map[uint32]uint64, *block.MetaBlock, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < e.shardCoordinator.NumberOfShards(); i++ {
		mapShardIdNonce[i] = e.genesisNonce
	}
	mapShardIdNonce[core.MetachainShardId] = e.genesisNonce

	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	previousEpochStartMeta, err := process.GetMetaHeaderFromStorage([]byte(epochStartIdentifier), e.marshalizer, e.store)
	if err != nil {
		return nil, nil, err
	}

	if epoch == e.genesisEpoch {
		return mapShardIdNonce, previousEpochStartMeta, nil
	}

	mapShardIdNonce[core.MetachainShardId] = previousEpochStartMeta.GetNonce()
	for _, shardData := range previousEpochStartMeta.EpochStart.LastFinalizedHeaders {
		mapShardIdNonce[shardData.ShardID] = shardData.Nonce
	}

	return mapShardIdNonce, previousEpochStartMeta, nil
}

func (e *economics) startNoncePerShardFromLastCrossNotarized(metaNonce uint64, epochStart block.EpochStart) (map[uint32]uint64, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < e.shardCoordinator.NumberOfShards(); i++ {
		mapShardIdNonce[i] = e.genesisNonce
	}
	mapShardIdNonce[core.MetachainShardId] = metaNonce

	for _, shardData := range epochStart.LastFinalizedHeaders {
		mapShardIdNonce[shardData.ShardID] = shardData.Nonce
	}

	return mapShardIdNonce, nil
}

// VerifyRewardsPerBlock checks whether rewards per block value was correctly computed
func (e *economics) VerifyRewardsPerBlock(metaBlock *block.MetaBlock, correctedProtocolSustainability *big.Int) error {
	if !metaBlock.IsStartOfEpochBlock() {
		return nil
	}
	computedEconomics, err := e.ComputeEndOfEpochEconomics(metaBlock)
	if err != nil {
		return err
	}
	computedEconomics.RewardsForProtocolSustainability.Set(correctedProtocolSustainability)
	computedEconomicsHash, err := core.CalculateHash(e.marshalizer, e.hasher, computedEconomics)
	if err != nil {
		return err
	}

	receivedEconomics := metaBlock.EpochStart.Economics
	receivedEconomicsHash, err := core.CalculateHash(e.marshalizer, e.hasher, &receivedEconomics)
	if err != nil {
		return err
	}

	if !bytes.Equal(receivedEconomicsHash, computedEconomicsHash) {
		logEconomicsDifferences(computedEconomics, &receivedEconomics)
		return epochStart.ErrEndOfEpochEconomicsDataDoesNotMatch
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (e *economics) IsInterfaceNil() bool {
	return e == nil
}

func logEconomicsDifferences(computed *block.Economics, received *block.Economics) {
	log.Warn("VerifyRewardsPerBlock error",
		"\ncomputed total to distribute", computed.TotalToDistribute,
		"computed total newly minted", computed.TotalNewlyMinted,
		"computed total supply", computed.TotalSupply,
		"computed rewards per block per node", computed.RewardsPerBlock,
		"computed rewards for protocol sustainability", computed.RewardsForProtocolSustainability,
		"computed node price", computed.NodePrice,
		"\nreceived total to distribute", received.TotalToDistribute,
		"received total newly minted", received.TotalNewlyMinted,
		"received total supply", received.TotalSupply,
		"received rewards per block per node", received.RewardsPerBlock,
		"received rewards for protocol sustainability", received.RewardsForProtocolSustainability,
		"received node price", received.NodePrice,
	)
}
