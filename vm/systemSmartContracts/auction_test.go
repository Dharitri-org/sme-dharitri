package systemSmartContracts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/process/smartContract/hooks"
	"github.com/Dharitri-org/sme-dharitri/vm"
	"github.com/Dharitri-org/sme-dharitri/vm/mock"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
	"github.com/Dharitri-org/sme-vm-common/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgumentsForAuction() ArgsStakingAuctionSmartContract {
	args := ArgsStakingAuctionSmartContract{
		Eei:                &mock.SystemEIStub{},
		SigVerifier:        &mock.MessageSignVerifierMock{},
		AuctionSCAddress:   []byte("auction"),
		StakingSCAddress:   []byte("staking"),
		NumOfNodesToSelect: 10,
		StakingSCConfig: config.StakingSystemSCConfig{
			GenesisNodePrice:                     "1000",
			UnJailValue:                          "10",
			MinStepValue:                         "10",
			MinStakeValue:                        "1",
			UnBondPeriod:                         1,
			AuctionEnableNonce:                   0,
			StakeEnableNonce:                     0,
			NumRoundsWithoutBleed:                1,
			MaximumPercentageToBleed:             1,
			BleedPercentagePerRound:              1,
			MaxNumberOfNodesForStake:             10,
			NodesToSelectInAuction:               100,
			ActivateBLSPubKeyMessageVerification: false,
		},
		Marshalizer:        &mock.MarshalizerMock{},
		GenesisTotalSupply: big.NewInt(100000000),
	}

	return args
}

func createABid(totalStakeValue uint64, numBlsKeys uint32, maxStakePerNode uint64) AuctionData {
	data := AuctionData{
		RewardAddress:   []byte("addr"),
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      nil,
		TotalStakeValue: big.NewInt(0).SetUint64(totalStakeValue),
		LockedStake:     big.NewInt(0).SetUint64(totalStakeValue),
		MaxStakePerNode: big.NewInt(0).SetUint64(maxStakePerNode),
	}

	keys := make([][]byte, 0)
	for i := uint32(0); i < numBlsKeys; i++ {
		keys = append(keys, []byte(fmt.Sprintf("%d", rand.Uint32())))
	}
	data.BlsPubKeys = keys

	return data
}

func TestNewStakingAuctionSmartContract_NilStakeValue(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForAuction()
	arguments.StakingSCConfig.GenesisNodePrice = ""

	asc, err := NewStakingAuctionSmartContract(arguments)
	require.Nil(t, asc)
	require.Equal(t, vm.ErrNegativeInitialStakeValue, err)
}

func TestNewStakingAuctionSmartContract_NegativeInitialStakeValue(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForAuction()
	arguments.StakingSCConfig.MinStepValue = ""

	asc, err := NewStakingAuctionSmartContract(arguments)
	require.Nil(t, asc)
	require.Equal(t, vm.ErrNegativeInitialStakeValue, err)
}

func TestNewStakingAuctionSmartContract_NilSystemEnvironmentInterface(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForAuction()
	arguments.Eei = nil

	asc, err := NewStakingAuctionSmartContract(arguments)
	require.Nil(t, asc)
	require.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewStakingAuctionSmartContract_NilStakingSmartContractAddress(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForAuction()
	arguments.StakingSCAddress = nil

	asc, err := NewStakingAuctionSmartContract(arguments)
	require.Nil(t, asc)
	require.Equal(t, vm.ErrNilStakingSmartContractAddress, err)
}

func TestNewStakingAuctionSmartContract_NilAuctionSmartContractAddress(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsForAuction()
	arguments.AuctionSCAddress = nil

	asc, err := NewStakingAuctionSmartContract(arguments)
	require.Nil(t, asc)
	require.Equal(t, vm.ErrNilAuctionSmartContractAddress, err)
}

func TestAuctionSC_calculateNodePrice_Case1(t *testing.T) {
	t.Parallel()

	expectedNodePrice := big.NewInt(20000000)
	args := createMockArgumentsForAuction()
	args.GenesisTotalSupply = big.NewInt(1000000000)
	args.StakingSCConfig.MinStakeValue = "100"
	stakingAuctionSC, _ := NewStakingAuctionSmartContract(args)

	bids := []AuctionData{
		createABid(100000000, 100, 30000000),
		createABid(20000000, 100, 20000000),
		createABid(20000000, 100, 20000000),
		createABid(20000000, 100, 20000000),
		createABid(20000000, 100, 20000000),
		createABid(20000000, 100, 20000000),
	}

	nodePrice, err := stakingAuctionSC.calculateNodePrice(bids)
	assert.Equal(t, expectedNodePrice, nodePrice)
	assert.Nil(t, err)
}

func TestAuctionSC_calculateNodePrice_Case2(t *testing.T) {
	t.Parallel()

	expectedNodePrice := big.NewInt(20000000)
	args := createMockArgumentsForAuction()
	args.NumOfNodesToSelect = 5
	stakingAuctionSC, _ := NewStakingAuctionSmartContract(args)

	bids := []AuctionData{
		createABid(100000000, 1, 30000000),
		createABid(50000000, 100, 25000000),
		createABid(30000000, 100, 15000000),
		createABid(40000000, 100, 20000000),
	}

	nodePrice, err := stakingAuctionSC.calculateNodePrice(bids)
	assert.Equal(t, expectedNodePrice, nodePrice)
	assert.Nil(t, err)
}

func TestAuctionSC_calculateNodePrice_Case3(t *testing.T) {
	t.Parallel()

	expectedNodePrice := big.NewInt(12500000)
	args := createMockArgumentsForAuction()
	args.NumOfNodesToSelect = 5
	stakingAuctionSC, _ := NewStakingAuctionSmartContract(args)

	bids := []AuctionData{
		createABid(25000000, 2, 12500000),
		createABid(30000000, 3, 10000000),
		createABid(40000000, 2, 20000000),
		createABid(50000000, 2, 25000000),
	}

	nodePrice, err := stakingAuctionSC.calculateNodePrice(bids)
	assert.Equal(t, expectedNodePrice, nodePrice)
	assert.Nil(t, err)
}

func TestAuctionSC_calculateNodePrice_Case4ShouldErr(t *testing.T) {
	t.Parallel()

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(createMockArgumentsForAuction())

	bid1 := createABid(25000000, 2, 12500000)
	bid2 := createABid(30000000, 3, 10000000)
	bid3 := createABid(40000000, 2, 20000000)
	bid4 := createABid(50000000, 2, 25000000)

	bids := []AuctionData{
		bid1, bid2, bid3, bid4,
	}

	nodePrice, err := stakingAuctionSC.calculateNodePrice(bids)
	assert.Nil(t, nodePrice)
	assert.Equal(t, vm.ErrNotEnoughQualifiedNodes, err)
}

func TestAuctionSC_selection_StakeGetAllocatedSeats(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForAuction()
	args.NumOfNodesToSelect = 5
	stakingAuctionSC, _ := NewStakingAuctionSmartContract(args)

	bid1 := createABid(25000000, 2, 12500000)
	bid2 := createABid(30000000, 3, 10000000)
	bid3 := createABid(40000000, 2, 20000000)
	bid4 := createABid(50000000, 2, 25000000)

	bids := []AuctionData{
		bid1, bid2, bid3, bid4,
	}

	// verify at least one is qualified from everybody
	expectedKeys := [][]byte{bid1.BlsPubKeys[0], bid3.BlsPubKeys[0], bid4.BlsPubKeys[0]}

	data := stakingAuctionSC.selection(bids)
	checkExpectedKeys(t, expectedKeys, data, len(expectedKeys))
}

func TestAuctionSC_selection_FirstBidderShouldTake50Percents(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForAuction()
	args.StakingSCConfig.MinStepValue = big.NewInt(100000).Text(10)
	stakingAuctionSC, _ := NewStakingAuctionSmartContract(args)

	bids := []AuctionData{
		createABid(10000000, 10, 10000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
	}

	data := stakingAuctionSC.selection(bids)
	//check that 50% keys belong to the first bidder
	checkExpectedKeys(t, bids[0].BlsPubKeys, data, 5)
}

func checkExpectedKeys(t *testing.T, expectedKeys [][]byte, data [][]byte, expectedNum int) {
	count := 0
	for _, key := range data {
		for _, expectedKey := range expectedKeys {
			if bytes.Equal(key, expectedKey) {
				count++
				break
			}
		}
	}
	assert.Equal(t, expectedNum, count)
}

func TestAuctionSC_selection_FirstBidderTakesAll(t *testing.T) {
	t.Parallel()

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(createMockArgumentsForAuction())

	bids := []AuctionData{
		createABid(100000000, 10, 10000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
		createABid(1000000, 1, 1000000),
	}

	data := stakingAuctionSC.selection(bids)
	//check that 100% keys belong to the first bidder
	checkExpectedKeys(t, bids[0].BlsPubKeys, data, 10)
}

func TestStakingAuctionSC_ExecuteStakeWithoutArgumentsShouldWork(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	auctionData := createABid(25000000, 2, 12500000)
	auctionDataBytes, _ := json.Marshal(&auctionData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Equal(key, arguments.CallerAddr) {
			return auctionDataBytes
		}
		return nil
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		if bytes.Equal(key, arguments.CallerAddr) {
			var auctionData AuctionData
			_ = json.Unmarshal(value, &auctionData)
			assert.Equal(t, big.NewInt(26000000), auctionData.TotalStakeValue)
		}
	}
	args := createMockArgumentsForAuction()
	args.Eei = eei
	args.NumOfNodesToSelect = 5

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(args)

	arguments.Function = "stake"
	arguments.CallValue = big.NewInt(1000000)

	errCode := stakingAuctionSC.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingAuctionSC_ExecuteStakeAddedNewPubKeysShouldWork(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	auctionData := createABid(25000000, 2, 12500000)
	auctionDataBytes, _ := json.Marshal(&auctionData)

	key1 := []byte("Key1")
	key2 := []byte("Key2")
	rewardAddr := []byte("tralala2")
	maxStakePerNoce := big.NewInt(500)

	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("auction"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.Eei = eei
	eei.SetStorage(arguments.CallerAddr, auctionDataBytes)
	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.NumOfNodesToSelect = 5

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(args)

	arguments.Function = "stake"
	arguments.CallValue = big.NewInt(0).Mul(big.NewInt(2), big.NewInt(10000000))
	arguments.Arguments = [][]byte{big.NewInt(2).Bytes(), key1, []byte("msg1"), key2, []byte("msg2"), maxStakePerNoce.Bytes(), rewardAddr}

	errCode := stakingAuctionSC.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingAuctionSC_ExecuteStakeWithRewardAddress(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("stakerAddress")
	stakerPubKey := []byte("stakerBLSPubKey")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig

	rwdAddress := []byte("rewardAddress")
	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed"), rwdAddress}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey)
	stakedData := &StakedData{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, bytes.Equal(stakedData.RewardAddress, rwdAddress))
}

func TestStakingAuctionSC_ExecuteStakeUnJail(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("address")
	stakerPubKey := []byte("blsPubKey")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	argsStaking.StakingSCConfig.UnJailValue = "1000"

	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig
	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unJail"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(1000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.CallerAddr = []byte("otherAddress")
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestStakingAuctionSC_ExecuteStakeUnStakeOneBlsPubKey(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	auctionData := createABid(25000000, 2, 12500000)
	auctionDataBytes, _ := json.Marshal(&auctionData)

	stakedData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 1,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: []byte("tralala1"),
		StakeValue:    big.NewInt(0),
	}
	stakedDataBytes, _ := json.Marshal(&stakedData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Equal(key, arguments.CallerAddr) {
			return auctionDataBytes
		}
		if bytes.Equal(key, auctionData.BlsPubKeys[0]) {
			return stakedDataBytes
		}
		return nil
	}
	eei.SetStorageCalled = func(key []byte, value []byte) {
		var stakedData StakedData
		_ = json.Unmarshal(value, &stakedData)

		assert.Equal(t, false, stakedData.Staked)
	}

	args := createMockArgumentsForAuction()
	args.Eei = eei
	args.NumOfNodesToSelect = 5

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(args)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{auctionData.BlsPubKeys[0]}
	errCode := stakingAuctionSC.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestStakingAuctionSC_ExecuteStakeStakeClaim(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "claim"
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	vmOutput := eei.CreateVMOutput()
	assert.NotNil(t, vmOutput)
	outputAccount := vmOutput.OutputAccounts[string(arguments.CallerAddr)]
	assert.True(t, outputAccount.BalanceDelta.Cmp(big.NewInt(10000000)) == 0)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey.Bytes())
	stakedData := &StakedData{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
}

func TestStakingAuctionSC_ExecuteStakeUnStakeOneBlsPubKeyAndRestake(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 100000
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.Eei = eei
	args.StakingSCConfig = argsStaking.StakingSCConfig
	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	addToStakeValue := big.NewInt(10000)
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = addToStakeValue
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "claim"
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	vmOutput := eei.CreateVMOutput()
	assert.NotNil(t, vmOutput)
	outputAccount := vmOutput.OutputAccounts[string(arguments.CallerAddr)]
	assert.True(t, outputAccount.BalanceDelta.Cmp(addToStakeValue) == 0)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey.Bytes())
	stakedData := &StakedData{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
}

func TestStakingAuctionSC_ExecuteStakeUnStakeUnBondUnStakeUnBondOneBlsPubKey(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("staker1")
	stakerPubKey := []byte("bls1")

	unBondPeriod := uint64(5)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = unBondPeriod
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.CallerAddr = []byte("anotherAddress")
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("anotherKey"), []byte("signed")}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 100
	}
	blockChainHook.CurrentEpochCalled = func() uint32 {
		return 10
	}

	arguments.CallerAddr = stakerAddress
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 103
	}
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 120
	}
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 220
	}
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return 320
	}
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey)
	assert.Equal(t, 0, len(marshaledData))
}

func TestStakingAuctionSC_ExecuteStakeChangeRewardAddresStakeUnStake(t *testing.T) {
	t.Parallel()

	stakerAddress := []byte("staker1")
	stakerPubKey := []byte("bls1")

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 1000
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	rewardAddress := []byte("staker2")
	arguments.Function = "changeRewardAddress"
	arguments.Arguments = [][]byte{rewardAddress}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	blsKey2 := []byte("blsKey2")
	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), blsKey2, []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{blsKey2}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey)
	stakedData := &StakedData{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
	assert.True(t, bytes.Equal(rewardAddress, stakedData.RewardAddress))

	marshaledData = eei.GetStorage(blsKey2)
	stakedData = &StakedData{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.False(t, stakedData.Staked)
	assert.True(t, bytes.Equal(rewardAddress, stakedData.RewardAddress))
}

func TestStakingAuctionSC_ExecuteStakeUnStakeUnBondBlsPubKeyAndRestake(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)
	nonce := uint64(1)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return nonce
		},
	}
	args := createMockArgumentsForAuction()

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = "10000000"
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = 1000
	argsStaking.StakingSCConfig.AuctionEnableNonce = 100000000
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.StakingSCConfig = argsStaking.StakingSCConfig
	args.Eei = eei

	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(10000000)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.CallerAddr = []byte("anotherCaller")
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), []byte("anotherKey"), []byte("signed")}
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	nonce += 1
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)

	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	nonce += args.StakingSCConfig.UnBondPeriod + 1
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{stakerPubKey.Bytes()}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	arguments.Function = "stake"
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(0)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	arguments.CallValue = big.NewInt(10000000)
	retCode = sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	marshaledData := eei.GetStorage(stakerPubKey.Bytes())
	stakedData := &StakedData{}
	_ = json.Unmarshal(marshaledData, stakedData)
	assert.True(t, stakedData.Staked)
}

func TestStakingAuctionSC_ExecuteUnBound(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	totalStake := uint64(25000000)

	auctionData := createABid(totalStake, 2, 12500000)
	auctionDataBytes, _ := json.Marshal(&auctionData)

	stakedData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 1,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: []byte("tralala1"),
		StakeValue:    big.NewInt(12500000),
	}
	stakedDataBytes, _ := json.Marshal(&stakedData)

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		if bytes.Equal(arguments.CallerAddr, key) {
			return auctionDataBytes
		}
		if bytes.Equal(key, auctionData.BlsPubKeys[0]) {
			return stakedDataBytes
		}

		return nil
	}

	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingAuctionSC, _ := NewStakingAuctionSmartContract(args)

	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{auctionData.BlsPubKeys[0]}
	errCode := stakingAuctionSC.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, errCode)
}

func TestAuctionStakingSC_ExecuteInit(t *testing.T) {
	t.Parallel()

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = core.SCDeployInitFunctionName

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	ownerAddr := stakingSmartContract.eei.GetStorage([]byte(ownerKey))
	assert.Equal(t, arguments.CallerAddr, ownerAddr)

	ownerBalanceBytes := stakingSmartContract.eei.GetStorage(arguments.CallerAddr)
	ownerBalance := big.NewInt(0).SetBytes(ownerBalanceBytes)
	assert.Equal(t, big.NewInt(0), ownerBalance)

}

func TestAuctionStakingSC_ExecuteInitTwoTimeShouldReturnUserError(t *testing.T) {
	t.Parallel()

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = core.SCDeployInitFunctionName

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteStakeWrongStakeValueShouldErr(t *testing.T) {
	t.Parallel()

	blockChainHook := &mock.BlockChainHookStub{
		GetUserAccountCalled: func(address []byte) (vmcommon.UserAccountHandler, error) {
			return nil, state.ErrAccNotFound
		},
	}
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{})
	eei.SetSCAddress([]byte("addr"))

	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	balance := eei.GetBalance(arguments.CallerAddr)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestAuctionStakingSC_ExecuteStakeWrongUnmarshalDataShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteStakeRegistrationDataStakedShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		registrationDataMarshalized, _ := json.Marshal(&StakedData{Staked: true})
		return registrationDataMarshalized
	}
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteStakeNotEnoughArgsShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		registrationDataMarshalized, _ := json.Marshal(&StakedData{})
		return registrationDataMarshalized
	}
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteStake(t *testing.T) {
	t.Parallel()

	stakerAddress := big.NewInt(100)
	stakerPubKey := big.NewInt(100)

	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	expectedRegistrationData := AuctionData{
		RewardAddress:   stakerAddress.Bytes(),
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{stakerPubKey.Bytes()},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.Eei = eei

	sc, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.CallerAddr = stakerAddress.Bytes()
	arguments.Arguments = [][]byte{big.NewInt(1).Bytes(), stakerPubKey.Bytes(), []byte("signed")}
	arguments.CallValue = big.NewInt(100).Set(nodePrice)

	retCode := sc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData AuctionData
	data := sc.eei.GetStorage(arguments.CallerAddr)
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestAuctionStakingSC_ExecuteUnStakeAddressNotStakedShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake@abc"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteUnStakeUnmarshalErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake@abc"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteUnStakeAlreadyUnStakedAddrShouldErr(t *testing.T) {
	t.Parallel()

	stakedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: nil,
		StakeValue:    nil,
	}

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes(), big.NewInt(200).Bytes()}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.CallerAddr, marshalizedExpectedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteUnStakeFailsWithWrongCaller(t *testing.T) {
	t.Parallel()

	expectedCallerAddress := []byte("caller")
	wrongCallerAddress := []byte("wrongCaller")

	stakedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: expectedCallerAddress,
		StakeValue:    nil,
	}

	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{})
	eei.SetSCAddress([]byte("addr"))
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{wrongCallerAddress}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteUnStake(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForAuction()
	args.StakingSCConfig.UnBondPeriod = 10
	callerAddress := []byte("caller")
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	expectedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nodePrice,
		JailedRound:   math.MaxUint64,
	}

	stakedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: callerAddress,
		StakeValue:    nil,
	}

	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSC, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	args.Eei = eei
	eei.SetSCAddress(args.AuctionSCAddress)

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "unStake"
	arguments.Arguments = [][]byte{[]byte("abc")}
	arguments.CallerAddr = callerAddress
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	stakingSmartContract.eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)

	auctionData := AuctionData{
		RewardAddress:   arguments.CallerAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: nodePrice,
		LockedStake:     nodePrice,
		MaxStakePerNode: nodePrice,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(auctionData)
	eei.SetStorage(arguments.CallerAddr, marshaledRegistrationData)

	stakedData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		UnStakedEpoch: core.DefaultUnstakedEpoch,
		RewardAddress: arguments.CallerAddr,
		StakeValue:    nodePrice,
		JailedRound:   math.MaxUint64,
	}
	marshaledStakedData, _ := json.Marshal(stakedData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshaledStakedData)
	stakingSC.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})
	eei.SetSCAddress(args.AuctionSCAddress)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedData
	eei.SetSCAddress(args.StakingSCAddress)
	data := eei.GetStorage(arguments.Arguments[0])
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)
	assert.Equal(t, expectedRegistrationData, registrationData)
}

func TestAuctionStakingSC_ExecuteUnBoundUnmarshalErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		return []byte("data")
	}
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes(), big.NewInt(200).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteUnBoundValidatorNotUnStakeShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.GetStorageCalled = func(key []byte) []byte {
		switch {
		case bytes.Equal(key, []byte(ownerKey)):
			return []byte("data")
		default:
			registrationDataMarshalized, _ := json.Marshal(&StakedData{UnStakedNonce: 0})
			return registrationDataMarshalized
		}
	}
	eei.BlockChainHookCalled = func() vmcommon.BlockchainHook {
		return &mock.BlockChainHookStub{CurrentNonceCalled: func() uint64 {
			return 10000
		}}
	}
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteStakeUnStakeReturnsErrAsNotEnabled(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	eei.BlockChainHookCalled = func() vmcommon.BlockchainHook {
		return &mock.BlockChainHookStub{CurrentNonceCalled: func() uint64 {
			return 100
		}}
	}
	args := createMockArgumentsForAuction()
	args.StakingSCConfig.StakeEnableNonce = eei.BlockChainHook().CurrentNonce() + uint64(1)
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{big.NewInt(100).Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	arguments.Function = "unStake"
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)

	arguments.Function = "stake"
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteUnBondBeforePeriodEnds(t *testing.T) {
	t.Parallel()

	unstakedNonce := uint64(10)
	registrationData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: unstakedNonce,
		RewardAddress: nil,
		StakeValue:    big.NewInt(100),
	}
	blsPubKey := big.NewInt(100)
	marshalizedRegData, _ := json.Marshal(&registrationData)
	eei, _ := NewVMContext(&mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return unstakedNonce + 1
		},
	},
		hooks.NewVMCryptoHook(),
		parsers.NewCallArgsParser(),
		&mock.AccountsStub{},
	)

	eei.SetSCAddress([]byte("addr"))
	eei.SetStorage([]byte(ownerKey), []byte("data"))
	eei.SetStorage(blsPubKey.Bytes(), marshalizedRegData)
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("data")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{blsPubKey.Bytes()}

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteUnBond(t *testing.T) {
	t.Parallel()

	unBondPeriod := uint64(100)
	unStakedNonce := uint64(10)
	stakeValue := big.NewInt(100)
	stakedData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: unStakedNonce,
		RewardAddress: []byte("reward"),
		StakeValue:    big.NewInt(0).Set(stakeValue),
		JailedRound:   math.MaxUint64,
	}

	marshalizedStakedData, _ := json.Marshal(&stakedData)
	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(&mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			return unStakedNonce + unBondPeriod + 1
		},
	}, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	scAddress := []byte("owner")
	eei.SetSCAddress(scAddress)
	eei.SetStorage([]byte(ownerKey), scAddress)

	args := createMockArgumentsForAuction()
	args.Eei = eei
	args.StakingSCConfig.GenesisNodePrice = stakeValue.Text(10)
	args.StakingSCConfig.UnBondPeriod = unBondPeriod
	args.StakingSCConfig.StakeEnableNonce = 0

	argsStaking := createMockStakingScArguments()
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig = args.StakingSCConfig
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.CallerAddr = []byte("address")
	arguments.Function = "unBond"
	arguments.Arguments = [][]byte{[]byte("abc")}
	arguments.RecipientAddr = scAddress

	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshalizedStakedData)
	stakingSC.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})
	eei.SetSCAddress(args.AuctionSCAddress)

	auctionData := AuctionData{
		RewardAddress:   arguments.CallerAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: stakeValue,
		LockedStake:     stakeValue,
		MaxStakePerNode: stakeValue,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(auctionData)
	eei.SetStorage(arguments.CallerAddr, marshaledRegistrationData)

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	eei.SetSCAddress(args.StakingSCAddress)
	data := eei.GetStorage(arguments.Arguments[0])
	assert.Equal(t, 0, len(data))

	destinationBalance := stakingSmartContract.eei.GetBalance(arguments.CallerAddr)
	scBalance := stakingSmartContract.eei.GetBalance(scAddress)
	assert.Equal(t, 0, destinationBalance.Cmp(stakeValue))
	assert.Equal(t, 0, scBalance.Cmp(big.NewInt(0).Mul(stakeValue, big.NewInt(-1))))
}

func TestAuctionStakingSC_ExecuteSlashOwnerAddrNotOkShouldErr(t *testing.T) {
	t.Parallel()

	eei := &mock.SystemEIStub{}
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	arguments := CreateVmContractCallInput()
	arguments.Function = "slash"

	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.UserError, retCode)
}

func TestAuctionStakingSC_ExecuteUnStakeAndUnBondStake(t *testing.T) {
	t.Parallel()

	// Preparation
	unBondPeriod := uint64(100)
	valueStakedByTheCaller := big.NewInt(100)
	stakerAddress := []byte("address")
	stakerPubKey := []byte("pubKey")
	blockChainHook := &mock.BlockChainHookStub{}
	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	smartcontractAddress := "auction"
	eei.SetSCAddress([]byte(smartcontractAddress))

	args := createMockArgumentsForAuction()
	args.Eei = eei
	args.StakingSCConfig.UnBondPeriod = unBondPeriod
	args.StakingSCConfig.GenesisNodePrice = valueStakedByTheCaller.Text(10)
	args.StakingSCConfig.AuctionEnableNonce = 0
	args.StakingSCConfig.StakeEnableNonce = 0

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig = args.StakingSCConfig
	argsStaking.Eei = eei
	stakingSC, _ := NewStakingSmartContract(argsStaking)
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)

	arguments := CreateVmContractCallInput()
	arguments.Arguments = [][]byte{stakerPubKey}
	arguments.CallerAddr = stakerAddress
	arguments.RecipientAddr = []byte(smartcontractAddress)

	stakedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        true,
		UnStakedNonce: 0,
		RewardAddress: stakerAddress,
		StakeValue:    valueStakedByTheCaller,
		JailedRound:   math.MaxUint64,
	}
	marshalizedExpectedRegData, _ := json.Marshal(&stakedRegistrationData)
	eei.SetSCAddress(args.StakingSCAddress)
	eei.SetStorage(arguments.Arguments[0], marshalizedExpectedRegData)
	stakingSC.setConfig(&StakingNodesConfig{MinNumNodes: 5, StakedNodes: 10})

	auctionData := AuctionData{
		RewardAddress:   arguments.CallerAddr,
		RegisterNonce:   0,
		Epoch:           0,
		BlsPubKeys:      [][]byte{arguments.Arguments[0]},
		TotalStakeValue: valueStakedByTheCaller,
		LockedStake:     valueStakedByTheCaller,
		MaxStakePerNode: valueStakedByTheCaller,
		NumRegistered:   1,
	}
	marshaledRegistrationData, _ := json.Marshal(auctionData)
	eei.SetSCAddress(args.AuctionSCAddress)
	eei.SetStorage(arguments.CallerAddr, marshaledRegistrationData)

	arguments.Function = "unStake"

	unStakeNonce := uint64(10)
	blockChainHook.CurrentNonceCalled = func() uint64 {
		return unStakeNonce
	}
	retCode := stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	var registrationData StakedData
	eei.SetSCAddress(args.StakingSCAddress)
	data := eei.GetStorage(arguments.Arguments[0])
	err := json.Unmarshal(data, &registrationData)
	assert.Nil(t, err)

	expectedRegistrationData := StakedData{
		RegisterNonce: 0,
		Staked:        false,
		UnStakedNonce: unStakeNonce,
		RewardAddress: stakerAddress,
		StakeValue:    valueStakedByTheCaller,
		JailedRound:   math.MaxUint64,
	}
	assert.Equal(t, expectedRegistrationData, registrationData)

	arguments.Function = "unBond"

	blockChainHook.CurrentNonceCalled = func() uint64 {
		return unStakeNonce + unBondPeriod + 1
	}
	eei.SetSCAddress(args.AuctionSCAddress)
	retCode = stakingSmartContract.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)

	destinationBalance := eei.GetBalance(arguments.CallerAddr)
	senderBalance := eei.GetBalance([]byte(smartcontractAddress))
	assert.Equal(t, big.NewInt(100), destinationBalance)
	assert.Equal(t, big.NewInt(-100), senderBalance)
}

func TestAuctionStakingSC_ExecuteGetShouldReturnUserErr(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	arguments.Function = "get"
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{})
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.UserError, err)
}

func TestAuctionStakingSC_ExecuteGetShouldOk(t *testing.T) {
	t.Parallel()

	arguments := CreateVmContractCallInput()
	arguments.Function = "get"
	arguments.Arguments = [][]byte{arguments.CallerAddr}
	eei, _ := NewVMContext(&mock.BlockChainHookStub{}, hooks.NewVMCryptoHook(), parsers.NewCallArgsParser(), &mock.AccountsStub{})
	args := createMockArgumentsForAuction()
	args.Eei = eei

	stakingSmartContract, _ := NewStakingAuctionSmartContract(args)
	err := stakingSmartContract.Execute(arguments)

	assert.Equal(t, vmcommon.Ok, err)
}

// Test scenario
// 1 -- will call claim from a account that does not stake -> will return error code
// 2 -- will do stake and lock all the stake value and claim should return error code because all the stake value is locked
// 3 -- will do stake and stake value will not be locked and after that claim should work
func TestAuctionStakingSC_Claim(t *testing.T) {
	t.Parallel()

	receiverAddr := []byte("receiverAddress")
	stakerAddress := []byte("stakerAddr")
	stakerPubKey := []byte("stakerPubKey")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	nodesToRunBytes := big.NewInt(1).Bytes()

	nonce := uint64(0)
	blockChainHook := &mock.BlockChainHookStub{
		CurrentNonceCalled: func() uint64 {
			defer func() {
				nonce++
			}()

			return nonce
		},
	}

	args := createMockArgumentsForAuction()
	args.Eei = createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)

	sc, _ := NewStakingAuctionSmartContract(args)

	//do claim should ret error
	doClaim(t, sc, stakerAddress, receiverAddr, vmcommon.UserError)

	//do stake
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	stake(t, sc, nodePrice, receiverAddr, stakerAddress, stakerPubKey, nodesToRunBytes)

	//do claim all stake is locked should return Ok
	doClaim(t, sc, stakerAddress, receiverAddr, vmcommon.Ok)

	// do stake to add more money but not lock the stake
	nonce = 0
	stake(t, sc, big.NewInt(1000), receiverAddr, stakerAddress, stakerPubKey, nodesToRunBytes)

	// do claim should work because not all the stake is locked
	doClaim(t, sc, stakerAddress, receiverAddr, vmcommon.Ok)
}

// Test scenario
// 1 -- call setConfig with wrong owner address should return error
// 2 -- call auction smart contract init and after that call setConfig with wrong number of arguments should return error
// 3 -- call setConfig after init was done successfully should work and config should be set correctly
func TestAuctionStakingSC_SetConfig(t *testing.T) {
	t.Parallel()

	ownerAddr := []byte("ownerAddress")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()
	args.Eei = createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)

	sc, _ := NewStakingAuctionSmartContract(args)

	// call setConfig should return error -> wrong owner address
	arguments := CreateVmContractCallInput()
	arguments.Function = "setConfig"
	retCode := sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)

	// call auction smart contract init
	arguments.Function = core.SCDeployInitFunctionName
	arguments.CallerAddr = ownerAddr
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.Ok, retCode)

	// call setConfig return error -> wrong number of arguments
	arguments.Function = "setConfig"
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.UserError, retCode)

	// call setConfig
	numNodes := big.NewInt(10)
	totalSupply := big.NewInt(10000000)
	minStep := big.NewInt(100)
	nodPrice := big.NewInt(20000)
	epoch := big.NewInt(1)
	unjailPrice := big.NewInt(100)
	arguments.Function = "setConfig"
	arguments.Arguments = [][]byte{minStakeValue.Bytes(), numNodes.Bytes(),
		totalSupply.Bytes(), minStep.Bytes(), nodPrice.Bytes(), unjailPrice.Bytes(), epoch.Bytes()}
	retCode = sc.Execute(arguments)
	require.Equal(t, vmcommon.Ok, retCode)

	auctionConfig := sc.getConfig(1)
	require.NotNil(t, auctionConfig)
	require.Equal(t, uint32(numNodes.Int64()), auctionConfig.NumNodes)
	require.Equal(t, totalSupply, auctionConfig.TotalSupply)
	require.Equal(t, minStep, auctionConfig.MinStep)
	require.Equal(t, nodPrice, auctionConfig.NodePrice)
	require.Equal(t, unjailPrice, auctionConfig.UnJailPrice)
	require.Equal(t, minStakeValue, auctionConfig.MinStakeValue)
}

func TestAuctionStakingSC_ChangeRewardAddress(t *testing.T) {
	t.Parallel()

	receiverAddr := []byte("receiverAddress")
	stakerAddress := []byte("stakerA")
	stakerPubKey := []byte("stakerP")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	nodesToRunBytes := big.NewInt(1).Bytes()
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()
	args.Eei = createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)

	sc, _ := NewStakingAuctionSmartContract(args)

	//change reward address should error nil arguments
	changeRewardAddress(t, sc, stakerAddress, nil, vmcommon.UserError)
	// change reward address should error wrong address
	changeRewardAddress(t, sc, stakerAddress, []byte("wrongAddress"), vmcommon.UserError)
	// change reward address should error because address is not belongs to any validator
	newRewardAddr := []byte("newAddr")
	changeRewardAddress(t, sc, stakerAddress, newRewardAddr, vmcommon.UserError)
	//do stake
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	stake(t, sc, nodePrice, receiverAddr, stakerAddress, stakerPubKey, nodesToRunBytes)

	// change reward address should error because new reward address is equal with old reward address
	changeRewardAddress(t, sc, stakerAddress, stakerAddress, vmcommon.UserError)
	// change reward address should work
	changeRewardAddress(t, sc, stakerAddress, newRewardAddr, vmcommon.Ok)
}

func TestAuctionStakingSC_ChangeValidatorKeys(t *testing.T) {
	t.Skip("function is disabled for now as it is not fully tested")
	t.Parallel()

	receiverAddr := []byte("receiverAddress")
	stakerAddress := []byte("stakerA")
	stakerPubKey := []byte("stakerP")
	minStakeValue := big.NewInt(1000)
	unboundPeriod := uint64(10)
	nodesToRunBytes := big.NewInt(1).Bytes()
	blockChainHook := &mock.BlockChainHookStub{}
	args := createMockArgumentsForAuction()
	args.Eei = createVmContextWithStakingSc(minStakeValue, unboundPeriod, blockChainHook)

	sc, _ := NewStakingAuctionSmartContract(args)

	// changeValidatorKeys should err not enough arguments
	newKey := []byte("newKey")
	changeValidatorKeys(t, sc, nodesToRunBytes, stakerAddress, stakerPubKey, newKey, nil, vmcommon.UserError)
	// changeValidatorKeys should error because address is not belongs to any validator
	changeValidatorKeys(t, sc, nodesToRunBytes, stakerAddress, stakerPubKey, newKey, []byte("signed"), vmcommon.UserError)
	//do stake
	nodePrice, _ := big.NewInt(0).SetString(args.StakingSCConfig.GenesisNodePrice, 10)
	stake(t, sc, nodePrice, receiverAddr, stakerAddress, stakerPubKey, nodesToRunBytes)
	// changeValidatorKeys should error not enough arguments
	nodesToRunBytes = big.NewInt(2).Bytes()
	changeValidatorKeys(t, sc, nodesToRunBytes, stakerAddress, stakerPubKey, newKey, []byte("signed"), vmcommon.UserError)
	// changeValidatorKeys should error verify sig will return error
	nodesToRunBytes = big.NewInt(1).Bytes()
	args.SigVerifier = &mock.MessageSignVerifierMock{
		VerifyCalled: func(message []byte, signedMessage []byte, pubKey []byte) error {
			return errors.New("new")
		},
	}
	changeValidatorKeys(t, sc, nodesToRunBytes, stakerAddress, stakerPubKey, newKey, []byte("signed"), vmcommon.UserError)
	// changeValidatorKeys should error wrong old key
	args.SigVerifier = &mock.MessageSignVerifierMock{}
	changeValidatorKeys(t, sc, nodesToRunBytes, stakerAddress, []byte("wrong"), newKey, []byte("signed"), vmcommon.UserError)

	// changeValidatorKeys should work
	newKey = []byte("newKey1")
	changeValidatorKeys(t, sc, nodesToRunBytes, stakerAddress, stakerPubKey, newKey, []byte("signed"), vmcommon.Ok)
}

func createVmContextWithStakingSc(stakeValue *big.Int, unboundPeriod uint64, blockChainHook vmcommon.BlockchainHook) *vmContext {
	atArgParser := parsers.NewCallArgsParser()
	eei, _ := NewVMContext(blockChainHook, hooks.NewVMCryptoHook(), atArgParser, &mock.AccountsStub{})

	argsStaking := createMockStakingScArguments()
	argsStaking.StakingSCConfig.GenesisNodePrice = stakeValue.Text(10)
	argsStaking.Eei = eei
	argsStaking.StakingSCConfig.UnBondPeriod = unboundPeriod
	stakingSC, _ := NewStakingSmartContract(argsStaking)

	eei.SetSCAddress([]byte("addr"))
	_ = eei.SetSystemSCContainer(&mock.SystemSCContainerStub{GetCalled: func(key []byte) (contract vm.SystemSmartContract, err error) {
		return stakingSC, nil
	}})

	return eei
}

func doClaim(t *testing.T, asc *stakingAuctionSC, stakerAddr, receiverAdd []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "claim"
	arguments.RecipientAddr = receiverAdd
	arguments.CallerAddr = stakerAddr

	retCode := asc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func stake(t *testing.T, asc *stakingAuctionSC, stakeValue *big.Int, receiverAdd, stakerAddr, stakerPubKey, nodesToRunBytes []byte) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "stake"
	arguments.RecipientAddr = receiverAdd
	arguments.CallerAddr = stakerAddr
	arguments.Arguments = [][]byte{nodesToRunBytes, stakerPubKey, []byte("signed")}
	arguments.CallValue = big.NewInt(0).Set(stakeValue)

	retCode := asc.Execute(arguments)
	assert.Equal(t, vmcommon.Ok, retCode)
}

func changeValidatorKeys(t *testing.T, asc *stakingAuctionSC, numNodes, stakedAddr, oldKey, newKey, signedMessage []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "changeValidatorKeys"
	arguments.CallerAddr = stakedAddr
	if signedMessage == nil {
		arguments.Arguments = nil
	} else {
		arguments.Arguments = [][]byte{numNodes, oldKey, newKey, signedMessage}
	}

	retCode := asc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}

func changeRewardAddress(t *testing.T, asc *stakingAuctionSC, callerAddr, newRewardAddr []byte, expectedCode vmcommon.ReturnCode) {
	arguments := CreateVmContractCallInput()
	arguments.Function = "changeRewardAddress"
	arguments.CallerAddr = callerAddr
	if newRewardAddr == nil {
		arguments.Arguments = nil
	} else {
		arguments.Arguments = [][]byte{newRewardAddr}
	}

	retCode := asc.Execute(arguments)
	assert.Equal(t, expectedCode, retCode)
}
