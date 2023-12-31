//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/Dharitri-org/protobuf/protobuf  --gogoslick_out=. staking.proto
package systemSmartContracts

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/vm"
	logger "github.com/Dharitri-org/sme-logger"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

var log = logger.GetOrCreate("vm/systemsmartcontracts")

const ownerKey = "owner"
const nodesConfigKey = "nodesConfig"
const waitingListHeadKey = "waitingList"
const waitingElementPrefix = "w_"

type stakingSC struct {
	eei                      vm.SystemEI
	stakeValue               *big.Int
	unBondPeriod             uint64
	stakeAccessAddr          []byte
	jailAccessAddr           []byte
	numRoundsWithoutBleed    uint64
	bleedPercentagePerRound  float64
	maximumPercentageToBleed float64
	gasCost                  vm.GasCost
	minNumNodes              uint64
	maxNumNodes              uint64
	marshalizer              marshal.Marshalizer
}

// ArgsNewStakingSmartContract holds the arguments needed to create a StakingSmartContract
type ArgsNewStakingSmartContract struct {
	StakingSCConfig   config.StakingSystemSCConfig
	MinNumNodes       uint64
	Eei               vm.SystemEI
	StakingAccessAddr []byte
	JailAccessAddr    []byte
	GasCost           vm.GasCost
	Marshalizer       marshal.Marshalizer
}

// NewStakingSmartContract creates a staking smart contract
func NewStakingSmartContract(
	args ArgsNewStakingSmartContract,
) (*stakingSC, error) {
	if check.IfNil(args.Eei) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if len(args.StakingAccessAddr) < 1 {
		return nil, vm.ErrInvalidStakingAccessAddress
	}
	if len(args.JailAccessAddr) < 1 {
		return nil, vm.ErrInvalidJailAccessAddress
	}
	if check.IfNil(args.Marshalizer) {
		return nil, vm.ErrNilMarshalizer
	}
	if args.StakingSCConfig.MaxNumberOfNodesForStake < 0 {
		return nil, vm.ErrNegativeWaitingNodesPercentage
	}
	if args.StakingSCConfig.BleedPercentagePerRound < 0 {
		return nil, vm.ErrNegativeBleedPercentagePerRound
	}
	if args.StakingSCConfig.MaximumPercentageToBleed < 0 {
		return nil, vm.ErrNegativeMaximumPercentageToBleed
	}
	if args.MinNumNodes > args.StakingSCConfig.MaxNumberOfNodesForStake {
		return nil, vm.ErrInvalidMaxNumberOfNodes
	}

	reg := &stakingSC{
		eei:                      args.Eei,
		unBondPeriod:             args.StakingSCConfig.UnBondPeriod,
		stakeAccessAddr:          args.StakingAccessAddr,
		jailAccessAddr:           args.JailAccessAddr,
		numRoundsWithoutBleed:    args.StakingSCConfig.NumRoundsWithoutBleed,
		bleedPercentagePerRound:  args.StakingSCConfig.BleedPercentagePerRound,
		maximumPercentageToBleed: args.StakingSCConfig.MaximumPercentageToBleed,
		gasCost:                  args.GasCost,
		minNumNodes:              args.MinNumNodes,
		maxNumNodes:              args.StakingSCConfig.MaxNumberOfNodesForStake,
		marshalizer:              args.Marshalizer,
	}
	conversionOk := true
	reg.stakeValue, conversionOk = big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, conversionBase)
	if !conversionOk || reg.stakeValue.Cmp(zero) < 0 {
		return nil, vm.ErrNegativeInitialStakeValue
	}

	return reg, nil
}

// Execute calls one of the functions from the staking smart contract and runs the code according to the input
func (r *stakingSC) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if CheckIfNil(args) != nil {
		return vmcommon.UserError
	}

	switch args.Function {
	case core.SCDeployInitFunctionName:
		return r.init(args)
	case "stake":
		return r.stake(args, false)
	case "register":
		return r.stake(args, true)
	case "unStake":
		return r.unStake(args)
	case "unBond":
		return r.unBond(args)
	case "slash":
		return r.slash(args)
	case "get":
		return r.get(args)
	case "isStaked":
		return r.isStaked(args)
	case "setStakeValue":
		return r.setStakeValueForCurrentEpoch(args)
	case "jail":
		return r.jail(args)
	case "unJail":
		return r.unJail(args)
	case "changeRewardAddress":
		return r.changeRewardAddress(args)
	case "changeValidatorKeys":
		return r.changeValidatorKey(args)
	}

	return vmcommon.UserError
}

func getPercentageOfValue(value *big.Int, percentage float64) *big.Int {
	x := new(big.Float).SetInt(value)
	y := big.NewFloat(percentage)

	z := new(big.Float).Mul(x, y)

	op := big.NewInt(0)
	result, _ := z.Int(op)

	return result
}

func (r *stakingSC) getConfig() *StakingNodesConfig {
	stakeConfig := &StakingNodesConfig{
		MinNumNodes: int64(r.minNumNodes),
		MaxNumNodes: int64(r.maxNumNodes),
	}
	configData := r.eei.GetStorage([]byte(nodesConfigKey))
	if len(configData) == 0 {
		return stakeConfig
	}

	err := r.marshalizer.Unmarshal(stakeConfig, configData)
	if err != nil {
		log.Warn("unmarshal error on getConfig function, returning baseConfig",
			"error", err.Error(),
		)
		return &StakingNodesConfig{
			MinNumNodes: int64(r.minNumNodes),
			MaxNumNodes: int64(r.maxNumNodes),
		}
	}

	return stakeConfig
}

func (r *stakingSC) setConfig(config *StakingNodesConfig) {
	configData, err := r.marshalizer.Marshal(config)
	if err != nil {
		log.Warn("marshal error on setConfig function",
			"error", err.Error(),
		)
		return
	}

	r.eei.SetStorage([]byte(nodesConfigKey), configData)
}

func (r *stakingSC) addToStakedNodes() {
	stakeConfig := r.getConfig()
	stakeConfig.StakedNodes++
	r.setConfig(stakeConfig)
}

func (r *stakingSC) removeFromStakedNodes() {
	stakeConfig := r.getConfig()
	if stakeConfig.StakedNodes > 0 {
		stakeConfig.StakedNodes--
	}
	r.setConfig(stakeConfig)
}

func (r *stakingSC) addToJailedNodes() {
	stakeConfig := r.getConfig()
	stakeConfig.JailedNodes++
	r.setConfig(stakeConfig)
}

func (r *stakingSC) removeFromJailedNodes() {
	stakeConfig := r.getConfig()
	if stakeConfig.JailedNodes > 0 {
		stakeConfig.JailedNodes--
	}
	r.setConfig(stakeConfig)
}

func (r *stakingSC) numSpareNodes() int64 {
	stakeConfig := r.getConfig()
	return stakeConfig.StakedNodes - stakeConfig.JailedNodes - stakeConfig.MinNumNodes
}

func (r *stakingSC) canStake() bool {
	stakeConfig := r.getConfig()
	return stakeConfig.StakedNodes < stakeConfig.MaxNumNodes
}

func (r *stakingSC) canUnStake() bool {
	return r.numSpareNodes() > 0
}

func (r *stakingSC) canUnBond() bool {
	return r.numSpareNodes() >= 0
}

func (r *stakingSC) calculateStakeAfterBleed(startRound uint64, endRound uint64, stake *big.Int) *big.Int {
	if startRound > endRound {
		return stake
	}
	if endRound-startRound < r.numRoundsWithoutBleed {
		return stake
	}

	totalRoundsToBleed := endRound - startRound - r.numRoundsWithoutBleed
	totalPercentageToBleed := float64(totalRoundsToBleed) * r.bleedPercentagePerRound
	totalPercentageToBleed = math.Min(totalPercentageToBleed, r.maximumPercentageToBleed)

	bleedValue := getPercentageOfValue(stake, totalPercentageToBleed)
	stakeAfterBleed := big.NewInt(0).Sub(stake, bleedValue)

	if stakeAfterBleed.Cmp(big.NewInt(0)) < 0 {
		stakeAfterBleed = big.NewInt(0)
	}

	return stakeAfterBleed
}

func (r *stakingSC) changeValidatorKey(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		r.eei.AddReturnMessage("changeValidatorKey function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		r.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 2, len(args.Arguments)))
		return vmcommon.UserError
	}

	oldKey := args.Arguments[0]
	newKey := args.Arguments[1]
	if len(oldKey) != len(newKey) {
		r.eei.AddReturnMessage("invalid bls key")
		return vmcommon.UserError
	}

	stakedData, err := r.getOrCreateRegisteredData(oldKey)
	if err != nil {
		r.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(stakedData.RewardAddress) == 0 {
		// if not registered this is not an error
		return vmcommon.Ok
	}

	r.eei.SetStorage(oldKey, nil)
	err = r.saveStakingData(newKey, stakedData)
	if err != nil {
		r.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (r *stakingSC) changeRewardAddress(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		r.eei.AddReturnMessage("stake function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		r.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 2, len(args.Arguments)))
		return vmcommon.UserError
	}

	newRewardAddress := args.Arguments[0]
	if len(newRewardAddress) != len(args.CallerAddr) {
		r.eei.AddReturnMessage("invalid reward address")
		return vmcommon.UserError
	}

	for _, blsKey := range args.Arguments[1:] {
		stakedData, err := r.getOrCreateRegisteredData(blsKey)
		if err != nil {
			r.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
			return vmcommon.UserError
		}
		if len(stakedData.RewardAddress) == 0 {
			continue
		}

		stakedData.RewardAddress = newRewardAddress
		err = r.saveStakingData(blsKey, stakedData)
		if err != nil {
			r.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (r *stakingSC) unJail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		r.eei.AddReturnMessage("unJail function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}

	for _, argument := range args.Arguments {
		stakedData, err := r.getOrCreateRegisteredData(argument)
		if err != nil {
			r.eei.AddReturnMessage("cannot get or created registered data: error " + err.Error())
			return vmcommon.UserError
		}
		if len(stakedData.RewardAddress) == 0 {
			r.eei.AddReturnMessage("cannot unJail a key that is not registered")
			return vmcommon.UserError
		}

		if stakedData.UnJailedNonce <= stakedData.JailedNonce {
			r.removeFromJailedNodes()
		}

		stakedData.StakeValue = r.calculateStakeAfterBleed(
			stakedData.JailedRound,
			r.eei.BlockChainHook().CurrentRound(),
			stakedData.StakeValue,
		)
		stakedData.JailedRound = math.MaxUint64
		stakedData.UnJailedNonce = r.eei.BlockChainHook().CurrentNonce()

		err = r.saveStakingData(argument, stakedData)
		if err != nil {
			r.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (r *stakingSC) getOrCreateRegisteredData(key []byte) (*StakedData, error) {
	registrationData := &StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: nil,
		StakeValue:    big.NewInt(0),
		JailedRound:   math.MaxUint64,
		UnJailedNonce: 0,
		JailedNonce:   0,
		StakedNonce:   math.MaxUint64,
	}

	data := r.eei.GetStorage(key)
	if len(data) > 0 {
		err := r.marshalizer.Unmarshal(registrationData, data)
		if err != nil {
			log.Debug("unmarshal error on staking SC stake function",
				"error", err.Error(),
			)
			return nil, err
		}
	}

	return registrationData, nil
}

func (r *stakingSC) saveStakingData(key []byte, stakedData *StakedData) error {
	data, err := r.marshalizer.Marshal(stakedData)
	if err != nil {
		log.Debug("marshal error on staking SC stake function ",
			"error", err.Error(),
		)
		return err
	}

	r.eei.SetStorage(key, data)
	return nil
}

func (r *stakingSC) jail(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.jailAccessAddr) {
		return vmcommon.UserError
	}

	for _, argument := range args.Arguments {
		stakedData, err := r.getOrCreateRegisteredData(argument)
		if err != nil {
			r.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
			return vmcommon.UserError
		}
		if len(stakedData.RewardAddress) == 0 {
			r.eei.AddReturnMessage("cannot jail a key that is not registered")
			return vmcommon.UserError
		}

		if stakedData.UnJailedNonce <= stakedData.JailedNonce {
			r.addToJailedNodes()
		}

		stakedData.JailedRound = r.eei.BlockChainHook().CurrentRound()
		stakedData.JailedNonce = r.eei.BlockChainHook().CurrentNonce()
		err = r.saveStakingData(argument, stakedData)
		if err != nil {
			r.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
			return vmcommon.UserError
		}
	}

	return vmcommon.Ok
}

func (r *stakingSC) get(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		r.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	value := r.eei.GetStorage(args.Arguments[0])
	r.eei.Finish(value)

	return vmcommon.Ok
}

func (r *stakingSC) init(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := r.eei.GetStorage([]byte(ownerKey))
	if ownerAddress != nil {
		r.eei.AddReturnMessage("smart contract was already initialized")
		return vmcommon.UserError
	}

	r.eei.SetStorage([]byte(ownerKey), args.CallerAddr)
	r.eei.SetStorage(args.CallerAddr, big.NewInt(0).Bytes())

	epoch := r.eei.BlockChainHook().CurrentEpoch()
	epochData := fmt.Sprintf("epoch_%d", epoch)

	r.eei.SetStorage([]byte(epochData), r.stakeValue.Bytes())

	stakeConfig := &StakingNodesConfig{
		MinNumNodes: int64(r.minNumNodes),
		MaxNumNodes: int64(r.maxNumNodes),
	}
	r.setConfig(stakeConfig)

	return vmcommon.Ok
}

func (r *stakingSC) setStakeValueForCurrentEpoch(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		r.eei.AddReturnMessage("setStakeValueForCurrentEpoch function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}

	if len(args.Arguments) < 1 {
		r.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments to call setStakeValueForCurrentEpoch: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	epoch := r.eei.BlockChainHook().CurrentEpoch()
	epochData := fmt.Sprintf("epoch_%d", epoch)

	inputStakeValue := big.NewInt(0).SetBytes(args.Arguments[0])
	if inputStakeValue.Cmp(r.stakeValue) < 0 {
		inputStakeValue.Set(r.stakeValue)
	}

	r.eei.SetStorage([]byte(epochData), inputStakeValue.Bytes())

	return vmcommon.Ok
}

func (r *stakingSC) getStakeValueForCurrentEpoch() *big.Int {
	stakeValue := big.NewInt(0)

	epoch := r.eei.BlockChainHook().CurrentEpoch()
	epochData := fmt.Sprintf("epoch_%d", epoch)

	stakeValueBytes := r.eei.GetStorage([]byte(epochData))
	stakeValue.SetBytes(stakeValueBytes)

	if stakeValue.Cmp(r.stakeValue) < 0 {
		stakeValue.Set(r.stakeValue)
	}

	return stakeValue
}

func (r *stakingSC) stake(args *vmcommon.ContractCallInput, onlyRegister bool) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		r.eei.AddReturnMessage("stake function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		r.eei.AddReturnMessage("not enough arguments, needed BLS key and reward address")
		return vmcommon.UserError
	}

	stakeValue := r.getStakeValueForCurrentEpoch()
	registrationData, err := r.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		r.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return vmcommon.UserError
	}

	if !onlyRegister {
		if registrationData.StakeValue.Cmp(stakeValue) < 0 {
			registrationData.StakeValue.Set(stakeValue)
		}

		if !r.canStake() && !registrationData.Staked {
			r.eei.AddReturnMessage("staking is full")
			err = r.addToWaitingList(args.Arguments[0])
			if err != nil {
				r.eei.AddReturnMessage("error while adding to waiting")
				return vmcommon.UserError
			}
			registrationData.Waiting = true
			r.eei.Finish([]byte{waiting})

			err = r.saveStakingData(args.Arguments[0], registrationData)
			if err != nil {
				r.eei.AddReturnMessage("cannot save staking registered data: error " + err.Error())
				return vmcommon.UserError
			}

			return vmcommon.Ok
		}

		if !registrationData.Staked {
			r.addToStakedNodes()
		}
		registrationData.Staked = true
		registrationData.StakedNonce = r.eei.BlockChainHook().CurrentNonce()
		registrationData.RegisterNonce = r.eei.BlockChainHook().CurrentNonce()
	}

	registrationData.RewardAddress = args.Arguments[1]

	err = r.saveStakingData(args.Arguments[0], registrationData)
	if err != nil {
		r.eei.AddReturnMessage("cannot save staking registered data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (r *stakingSC) unStake(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		r.eei.AddReturnMessage("unStake function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) < 2 {
		r.eei.AddReturnMessage("not enough arguments, needed BLS key and reward address")
		return vmcommon.UserError
	}

	registrationData, err := r.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		r.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		r.eei.AddReturnMessage("cannot unStake a key that is not registered")
		return vmcommon.UserError
	}

	if !bytes.Equal(args.Arguments[1], registrationData.RewardAddress) {
		r.eei.AddReturnMessage("unStake possible only from staker caller")
		return vmcommon.UserError
	}

	if !registrationData.Staked {
		r.eei.AddReturnMessage("unStake is not possible for address with is already unStaked")
		return vmcommon.UserError
	}
	if registrationData.JailedRound != math.MaxUint64 {
		r.eei.AddReturnMessage("unStake is not possible for jailed nodes")
		return vmcommon.UserError
	}
	err = r.moveFirstFromWaitingToStakedIfNeeded(args.Arguments[0])
	if err != nil {
		r.eei.AddReturnMessage(err.Error())
		return vmcommon.UserError
	}
	if !r.canUnStake() {
		r.eei.AddReturnMessage("unStake is not possible as too many left")
		return vmcommon.UserError
	}

	r.removeFromStakedNodes()
	registrationData.Staked = false
	registrationData.UnStakedEpoch = r.eei.BlockChainHook().CurrentEpoch()
	registrationData.UnStakedNonce = r.eei.BlockChainHook().CurrentNonce()

	err = r.saveStakingData(args.Arguments[0], registrationData)
	if err != nil {
		r.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (r *stakingSC) moveFirstFromWaitingToStakedIfNeeded(blsKey []byte) error {
	waitingElementKey := r.createWaitingListKey(blsKey)
	elementInList, err := r.getWaitingListElement(waitingElementKey)
	if err == nil {
		// node in waiting - remove from it - and that's it
		return r.removeFromWaitingList(blsKey)
	}

	waitingList, err := r.getWaitingListHead()
	if err != nil {
		return err
	}
	if waitingList.Length == 0 {
		return nil
	}
	elementInList, err = r.getWaitingListElement(waitingList.FirstKey)
	if err != nil {
		return err
	}
	err = r.removeFromWaitingList(elementInList.BLSPublicKey)
	if err != nil {
		return err
	}

	nodeData, err := r.getOrCreateRegisteredData(elementInList.BLSPublicKey)
	if err != nil {
		return err
	}
	if len(nodeData.RewardAddress) == 0 || nodeData.Staked {
		return vm.ErrInvalidWaitingList
	}

	nodeData.Waiting = false
	nodeData.Staked = true
	nodeData.RegisterNonce = r.eei.BlockChainHook().CurrentNonce()
	nodeData.StakedNonce = r.eei.BlockChainHook().CurrentNonce()
	stakeValue := r.getStakeValueForCurrentEpoch()
	if nodeData.StakeValue.Cmp(stakeValue) < 0 {
		nodeData.StakeValue.Set(stakeValue)
	}

	r.addToStakedNodes()
	return r.saveStakingData(elementInList.BLSPublicKey, nodeData)
}

func (r *stakingSC) unBond(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if !bytes.Equal(args.CallerAddr, r.stakeAccessAddr) {
		r.eei.AddReturnMessage("unBond function not allowed to be called by address " + string(args.CallerAddr))
		return vmcommon.UserError
	}
	if len(args.Arguments) < 1 {
		r.eei.AddReturnMessage("not enough arguments, needed BLS key")
		return vmcommon.UserError
	}

	registrationData, err := r.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		r.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return vmcommon.UserError
	}

	blsKeyBytes := make([]byte, 2*len(args.Arguments[0]))
	_ = hex.Encode(blsKeyBytes, args.Arguments[0])
	encodedBlsKey := string(blsKeyBytes)

	if len(registrationData.RewardAddress) == 0 {
		r.eei.AddReturnMessage(fmt.Sprintf("cannot unBond key %s that is not registered", encodedBlsKey))
		return vmcommon.UserError
	}
	if registrationData.Staked {
		r.eei.AddReturnMessage(fmt.Sprintf("unBond is not possible for key %s which is staked", encodedBlsKey))
		return vmcommon.UserError
	}
	if registrationData.JailedRound != math.MaxUint64 {
		r.eei.AddReturnMessage(fmt.Sprintf("unBond is not possible for jailed key %s", encodedBlsKey))
		return vmcommon.UserError
	}

	if r.isInWaiting(args.Arguments[0]) || registrationData.Waiting {
		err = r.removeFromWaitingList(args.Arguments[0])
		if err != nil {
			r.eei.AddReturnMessage(err.Error())
			return vmcommon.UserError
		}

		r.eei.SetStorage(args.Arguments[0], nil)
		r.eei.Finish(registrationData.StakeValue.Bytes())
		r.eei.Finish(big.NewInt(0).SetUint64(uint64(registrationData.UnStakedEpoch)).Bytes())

		return vmcommon.Ok
	}

	currentNonce := r.eei.BlockChainHook().CurrentNonce()
	if currentNonce-registrationData.UnStakedNonce < r.unBondPeriod {
		r.eei.AddReturnMessage(fmt.Sprintf("unBond is not possible for key %s because unBond period did not pass", encodedBlsKey))
		return vmcommon.UserError
	}
	if !r.canUnBond() {
		r.eei.AddReturnMessage("unbonding currently unavailable: number of total validators in the network is at minimum")
		return vmcommon.UserError
	}
	if r.eei.IsValidator(args.Arguments[0]) {
		r.eei.AddReturnMessage("unbonding is not possible: the node with key " + encodedBlsKey + " is still a validator")
		return vmcommon.UserError
	}

	r.eei.SetStorage(args.Arguments[0], nil)
	r.eei.Finish(registrationData.StakeValue.Bytes())
	r.eei.Finish(big.NewInt(0).SetUint64(uint64(registrationData.UnStakedEpoch)).Bytes())

	return vmcommon.Ok
}

func (r *stakingSC) slash(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	ownerAddress := r.eei.GetStorage([]byte(ownerKey))
	if !bytes.Equal(ownerAddress, args.CallerAddr) {
		r.eei.AddReturnMessage("slash function called by not the owners address")
		return vmcommon.UserError
	}

	if len(args.Arguments) != 2 {
		retMessage := fmt.Sprintf("slash function called with wrong number of arguments: expected %d, got %d", 2, len(args.Arguments))
		r.eei.AddReturnMessage(retMessage)
		return vmcommon.UserError
	}

	registrationData, err := r.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		r.eei.AddReturnMessage("cannot get ore create registered data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		r.eei.AddReturnMessage("invalid reward address")
		return vmcommon.UserError
	}
	if !registrationData.Staked {
		r.eei.AddReturnMessage("cannot slash already unstaked or user not staked")
		return vmcommon.UserError
	}

	if registrationData.UnJailedNonce >= registrationData.JailedNonce {
		r.addToJailedNodes()
	}

	stakedValue := big.NewInt(0).Set(registrationData.StakeValue)
	slashValue := big.NewInt(0).SetBytes(args.Arguments[1])
	registrationData.StakeValue = registrationData.StakeValue.Sub(stakedValue, slashValue)
	registrationData.JailedRound = r.eei.BlockChainHook().CurrentRound()
	registrationData.JailedNonce = r.eei.BlockChainHook().CurrentNonce()

	err = r.saveStakingData(args.Arguments[0], registrationData)
	if err != nil {
		r.eei.AddReturnMessage("cannot save staking data: error " + err.Error())
		return vmcommon.UserError
	}

	return vmcommon.Ok
}

func (r *stakingSC) isStaked(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if len(args.Arguments) < 1 {
		r.eei.AddReturnMessage(fmt.Sprintf("invalid number of arguments: expected min %d, got %d", 1, 0))
		return vmcommon.UserError
	}

	registrationData, err := r.getOrCreateRegisteredData(args.Arguments[0])
	if err != nil {
		r.eei.AddReturnMessage("cannot get or create registered data: error " + err.Error())
		return vmcommon.UserError
	}
	if len(registrationData.RewardAddress) == 0 {
		r.eei.AddReturnMessage("key is not registered")
		return vmcommon.UserError
	}

	if registrationData.Staked {
		return vmcommon.Ok
	}

	r.eei.AddReturnMessage("account not staked")
	return vmcommon.UserError
}

func (r *stakingSC) addToWaitingList(blsKey []byte) error {
	inWaitingListKey := r.createWaitingListKey(blsKey)
	marshaledData := r.eei.GetStorage(inWaitingListKey)
	if len(marshaledData) != 0 {
		return nil
	}

	waitingList, err := r.getWaitingListHead()
	if err != nil {
		return err
	}

	waitingList.Length += 1
	if waitingList.Length == 1 {
		waitingList.FirstKey = inWaitingListKey
		waitingList.LastKey = inWaitingListKey
		elementInWaiting := &ElementInList{
			BLSPublicKey: blsKey,
			PreviousKey:  waitingList.LastKey,
			NextKey:      make([]byte, 0),
		}
		return r.saveElementAndList(inWaitingListKey, elementInWaiting, waitingList)
	}

	oldLastKey := make([]byte, len(waitingList.LastKey))
	copy(oldLastKey, waitingList.LastKey)

	lastElement, err := r.getWaitingListElement(waitingList.LastKey)
	if err != nil {
		return err
	}
	lastElement.NextKey = inWaitingListKey
	elementInWaiting := &ElementInList{
		BLSPublicKey: blsKey,
		PreviousKey:  oldLastKey,
		NextKey:      make([]byte, 0),
	}

	err = r.saveWaitingListElement(oldLastKey, lastElement)
	if err != nil {
		return err
	}

	waitingList.LastKey = inWaitingListKey
	return r.saveElementAndList(inWaitingListKey, elementInWaiting, waitingList)
}

func (r *stakingSC) saveElementAndList(key []byte, element *ElementInList, waitingList *WaitingList) error {
	err := r.saveWaitingListElement(key, element)
	if err != nil {
		return err
	}

	return r.saveWaitingListHead(waitingList)
}

func (r *stakingSC) removeFromWaitingList(blsKey []byte) error {
	inWaitingListKey := r.createWaitingListKey(blsKey)
	marshaledData := r.eei.GetStorage(inWaitingListKey)
	if len(marshaledData) == 0 {
		return nil
	}
	r.eei.SetStorage(inWaitingListKey, nil)

	elementToRemove := &ElementInList{}
	err := r.marshalizer.Unmarshal(elementToRemove, marshaledData)
	if err != nil {
		return err
	}

	waitingList, err := r.getWaitingListHead()
	if err != nil {
		return err
	}
	if waitingList.Length == 0 {
		return vm.ErrInvalidWaitingList
	}
	waitingList.Length -= 1
	if waitingList.Length == 0 {
		r.eei.SetStorage([]byte(waitingListHeadKey), nil)
		return nil
	}
	if bytes.Equal(elementToRemove.PreviousKey, inWaitingListKey) {
		nextElement, err := r.getWaitingListElement(elementToRemove.NextKey)
		if err != nil {
			return err
		}

		nextElement.PreviousKey = elementToRemove.NextKey
		waitingList.FirstKey = elementToRemove.NextKey
		return r.saveElementAndList(elementToRemove.NextKey, nextElement, waitingList)
	}

	previousElement, err := r.getWaitingListElement(elementToRemove.PreviousKey)
	if err != nil {
		return err
	}
	if len(elementToRemove.NextKey) == 0 {
		waitingList.LastKey = elementToRemove.PreviousKey
		previousElement.NextKey = make([]byte, 0)
		return r.saveElementAndList(elementToRemove.PreviousKey, previousElement, waitingList)
	}

	nextElement, err := r.getWaitingListElement(elementToRemove.NextKey)
	if err != nil {
		return err
	}

	nextElement.PreviousKey = elementToRemove.PreviousKey
	previousElement.NextKey = elementToRemove.NextKey

	err = r.saveWaitingListElement(elementToRemove.NextKey, nextElement)
	if err != nil {
		return err
	}
	return r.saveElementAndList(elementToRemove.PreviousKey, previousElement, waitingList)
}

func (r *stakingSC) getWaitingListElement(key []byte) (*ElementInList, error) {
	marshaledData := r.eei.GetStorage(key)
	if len(marshaledData) == 0 {
		return nil, vm.ErrElementNotFound
	}

	element := &ElementInList{}
	err := r.marshalizer.Unmarshal(element, marshaledData)
	if err != nil {
		return nil, err
	}

	return element, nil
}

func (r *stakingSC) isInWaiting(blsKey []byte) bool {
	waitingKey := r.createWaitingListKey(blsKey)
	marshaledData := r.eei.GetStorage(waitingKey)
	return len(marshaledData) > 0
}

func (r *stakingSC) saveWaitingListElement(key []byte, element *ElementInList) error {
	marshaledData, err := r.marshalizer.Marshal(element)
	if err != nil {
		return err
	}

	r.eei.SetStorage(key, marshaledData)
	return nil
}

func (r *stakingSC) getWaitingListHead() (*WaitingList, error) {
	waitingList := &WaitingList{
		FirstKey: make([]byte, 0),
		LastKey:  make([]byte, 0),
		Length:   0,
	}
	marshaledData := r.eei.GetStorage([]byte(waitingListHeadKey))
	if len(marshaledData) == 0 {
		return waitingList, nil
	}

	err := r.marshalizer.Unmarshal(waitingList, marshaledData)
	if err != nil {
		return nil, err
	}

	return waitingList, nil
}

func (r *stakingSC) saveWaitingListHead(waitingList *WaitingList) error {
	marshaledData, err := r.marshalizer.Marshal(waitingList)
	if err != nil {
		return err
	}

	r.eei.SetStorage([]byte(waitingListHeadKey), marshaledData)
	return nil
}

func (r *stakingSC) createWaitingListKey(blsKey []byte) []byte {
	return []byte(waitingElementPrefix + string(blsKey))
}

// IsInterfaceNil verifies if the underlying object is nil or not
func (r *stakingSC) IsInterfaceNil() bool {
	return r == nil
}
