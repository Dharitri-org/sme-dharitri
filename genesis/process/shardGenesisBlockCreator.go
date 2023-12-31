package process

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	dataBlock "github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/genesis"
	"github.com/Dharitri-org/sme-dharitri/genesis/process/disabled"
	"github.com/Dharitri-org/sme-dharitri/genesis/process/intermediate"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/block/preprocess"
	"github.com/Dharitri-org/sme-dharitri/process/coordinator"
	"github.com/Dharitri-org/sme-dharitri/process/factory/shard"
	"github.com/Dharitri-org/sme-dharitri/process/rewardTransaction"
	"github.com/Dharitri-org/sme-dharitri/process/smartContract"
	"github.com/Dharitri-org/sme-dharitri/process/smartContract/builtInFunctions"
	"github.com/Dharitri-org/sme-dharitri/process/smartContract/hooks"
	"github.com/Dharitri-org/sme-dharitri/process/transaction"
	hardForkProcess "github.com/Dharitri-org/sme-dharitri/update/process"
	logger "github.com/Dharitri-org/sme-logger"
	"github.com/Dharitri-org/sme-vm-common/parsers"
)

var log = logger.GetOrCreate("genesis/process")

var zero = big.NewInt(0)

type deployedScMetrics struct {
	numDelegation int
	numOtherTypes int
}

// CreateShardGenesisBlock will create a shard genesis block
func CreateShardGenesisBlock(arg ArgsGenesisBlockCreator, nodesListSplitter genesis.NodesListSplitter, selfShardID uint32) (data.HeaderHandler, error) {
	if mustDoHardForkImportProcess(arg) {
		return createShardGenesisAfterHardFork(arg, selfShardID)
	}

	processors, err := createProcessorsForShard(arg)
	if err != nil {
		return nil, err
	}

	deployMetrics := &deployedScMetrics{}

	err = deployInitialSmartContracts(processors, arg, deployMetrics)
	if err != nil {
		return nil, err
	}

	numSetBalances, err := setBalancesToTrie(arg)
	if err != nil {
		return nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while setting the balances to trie",
			err, arg.ShardCoordinator.SelfId())
	}

	numStaked, err := increaseStakersNonces(processors, arg)
	if err != nil {
		return nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while incrementing nonces",
			err, arg.ShardCoordinator.SelfId())
	}

	delegationResult, err := executeDelegation(processors, arg, nodesListSplitter)
	if err != nil {
		return nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while execution delegation",
			err, arg.ShardCoordinator.SelfId())
	}

	numCrossShardDelegations, err := incrementNoncesForCrossShardDelegations(processors, arg)
	if err != nil {
		return nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while incrementing crossshard nonce",
			err, arg.ShardCoordinator.SelfId())
	}

	rootHash, err := arg.Accounts.Commit()
	if err != nil {
		return nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while commiting",
			err, arg.ShardCoordinator.SelfId())
	}

	log.Debug("shard block genesis",
		"shard ID", arg.ShardCoordinator.SelfId(),
		"num delegation SC deployed", deployMetrics.numDelegation,
		"num other SC deployed", deployMetrics.numOtherTypes,
		"num set balances", numSetBalances,
		"num staked directly", numStaked,
		"total staked on a delegation SC", delegationResult.NumTotalStaked,
		"total delegation nodes", delegationResult.NumTotalDelegated,
		"cross shard delegation calls", numCrossShardDelegations,
	)

	round, nonce, epoch := getGenesisBlocksRoundNonceEpoch(arg)
	header := &block.Header{
		Epoch:           epoch,
		Round:           round,
		Nonce:           nonce,
		ShardID:         arg.ShardCoordinator.SelfId(),
		BlockBodyType:   block.StateBlock,
		PubKeysBitmap:   []byte{1},
		Signature:       rootHash,
		RootHash:        rootHash,
		PrevRandSeed:    rootHash,
		RandSeed:        rootHash,
		TimeStamp:       arg.GenesisTime,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		ChainID:         []byte(arg.ChainID),
		SoftwareVersion: []byte(""),
	}

	err = processors.vmContainer.Close()
	if err != nil {
		return nil, err
	}

	return header, nil
}

func createShardGenesisAfterHardFork(arg ArgsGenesisBlockCreator, selfShardId uint32) (data.HeaderHandler, error) {
	tmpArg := arg
	tmpArg.Accounts = arg.importHandler.GetAccountsDBForShard(arg.ShardCoordinator.SelfId())
	processors, err := createProcessorsForShard(tmpArg)
	if err != nil {
		return nil, err
	}

	argsPendingTxProcessor := hardForkProcess.ArgsPendingTransactionProcessor{
		Accounts:         tmpArg.Accounts,
		TxProcessor:      processors.txProcessor,
		RwdTxProcessor:   processors.rwdProcessor,
		ScrTxProcessor:   processors.scrProcessor,
		PubKeyConv:       arg.PubkeyConv,
		ShardCoordinator: arg.ShardCoordinator,
	}
	pendingTxProcessor, err := hardForkProcess.NewPendingTransactionProcessor(argsPendingTxProcessor)
	if err != nil {
		return nil, err
	}

	argsShardBlockAfterHardFork := hardForkProcess.ArgsNewShardBlockCreatorAfterHardFork{
		ShardCoordinator:   arg.ShardCoordinator,
		TxCoordinator:      processors.txCoordinator,
		PendingTxProcessor: pendingTxProcessor,
		ImportHandler:      arg.importHandler,
		Marshalizer:        arg.Marshalizer,
		Hasher:             arg.Hasher,
		DataPool:           arg.DataPool,
		Storage:            arg.Store,
		SelfShardID:        selfShardId,
	}
	shardBlockCreator, err := hardForkProcess.NewShardBlockCreatorAfterHardFork(argsShardBlockAfterHardFork)
	if err != nil {
		return nil, err
	}

	hdrHandler, _, err := shardBlockCreator.CreateNewBlock(
		arg.ChainID,
		arg.HardForkConfig.StartRound,
		arg.HardForkConfig.StartNonce,
		arg.HardForkConfig.StartEpoch,
	)
	if err != nil {
		return nil, err
	}
	hdrHandler.SetTimeStamp(arg.GenesisTime)

	err = arg.Accounts.RecreateTrie(hdrHandler.GetRootHash())
	if err != nil {
		return nil, err
	}

	return hdrHandler, nil
}

// setBalancesToTrie adds balances to trie
func setBalancesToTrie(arg ArgsGenesisBlockCreator) (int, error) {
	initialAccounts, err := arg.AccountsParser.InitialAccountsSplitOnAddressesShards(arg.ShardCoordinator)
	if err != nil {
		return 0, err
	}

	initialAccountsForShard := initialAccounts[arg.ShardCoordinator.SelfId()]

	for _, accnt := range initialAccountsForShard {
		err = setBalanceToTrie(arg, accnt)
		if err != nil {
			return 0, err
		}
	}

	return len(initialAccountsForShard), nil
}

func setBalanceToTrie(arg ArgsGenesisBlockCreator, accnt genesis.InitialAccountHandler) error {
	accWrp, err := arg.Accounts.LoadAccount(accnt.AddressBytes())
	if err != nil {
		return err
	}

	account, ok := accWrp.(state.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = account.AddToBalance(accnt.GetBalanceValue())
	if err != nil {
		return err
	}

	return arg.Accounts.SaveAccount(account)
}

func createProcessorsForShard(arg ArgsGenesisBlockCreator) (*genesisProcessors, error) {
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:               arg.GasMap,
		MapDNSAddresses:      make(map[string]struct{}),
		EnableUserNameChange: false,
		Marshalizer:          arg.Marshalizer,
	}
	builtInFuncs, err := builtInFunctions.CreateBuiltInFunctionContainer(argsBuiltIn)
	if err != nil {
		return nil, err
	}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         arg.Accounts,
		PubkeyConv:       arg.PubkeyConv,
		StorageService:   arg.Store,
		BlockChain:       arg.Blkc,
		ShardCoordinator: arg.ShardCoordinator,
		Marshalizer:      arg.Marshalizer,
		Uint64Converter:  arg.Uint64ByteSliceConverter,
		BuiltInFunctions: builtInFuncs,
	}
	vmFactoryImpl, err := shard.NewVMContainerFactory(
		arg.VirtualMachineConfig,
		math.MaxUint64,
		arg.GasMap,
		argsHook,
	)
	if err != nil {
		return nil, err
	}

	vmContainer, err := vmFactoryImpl.Create()
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := shard.NewIntermediateProcessorsContainerFactory(
		arg.ShardCoordinator,
		arg.Marshalizer,
		arg.Hasher,
		arg.PubkeyConv,
		arg.Store,
		arg.DataPool,
	)
	if err != nil {
		return nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, err
	}

	scForwarder, err := interimProcContainer.Get(dataBlock.SmartContractResultBlock)
	if err != nil {
		return nil, err
	}

	receiptTxInterim, err := interimProcContainer.Get(dataBlock.ReceiptBlock)
	if err != nil {
		return nil, err
	}

	badTxInterim, err := interimProcContainer.Get(dataBlock.InvalidBlock)
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  arg.PubkeyConv,
		ShardCoordinator: arg.ShardCoordinator,
		BuiltInFuncNames: builtInFuncs.Keys(),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(arg.Economics, txTypeHandler)
	if err != nil {
		return nil, err
	}

	genesisFeeHandler := &disabled.FeeHandler{}
	argsNewScProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:      vmContainer,
		ArgsParser:       smartContract.NewArgumentParser(),
		Hasher:           arg.Hasher,
		Marshalizer:      arg.Marshalizer,
		AccountsDB:       arg.Accounts,
		BlockChainHook:   vmFactoryImpl.BlockChainHookImpl(),
		PubkeyConv:       arg.PubkeyConv,
		Coordinator:      arg.ShardCoordinator,
		ScrForwarder:     scForwarder,
		TxFeeHandler:     genesisFeeHandler,
		EconomicsFee:     genesisFeeHandler,
		TxTypeHandler:    txTypeHandler,
		GasHandler:       gasHandler,
		BuiltInFunctions: vmFactoryImpl.BlockChainHookImpl().GetBuiltInFunctions(),
		TxLogsProcessor:  arg.TxLogsProcessor,
		BadTxForwarder:   badTxInterim,
	}
	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewScProcessor)
	if err != nil {
		return nil, err
	}

	rewardsTxProcessor, err := rewardTransaction.NewRewardTxProcessor(
		arg.Accounts,
		arg.PubkeyConv,
		arg.ShardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	argsNewTxProcessor := transaction.ArgsNewTxProcessor{
		Accounts:          arg.Accounts,
		Hasher:            arg.Hasher,
		PubkeyConv:        arg.PubkeyConv,
		Marshalizer:       arg.Marshalizer,
		SignMarshalizer:   arg.SignMarshalizer,
		ShardCoordinator:  arg.ShardCoordinator,
		ScProcessor:       scProcessor,
		TxFeeHandler:      genesisFeeHandler,
		TxTypeHandler:     txTypeHandler,
		EconomicsFee:      genesisFeeHandler,
		ReceiptForwarder:  receiptTxInterim,
		BadTxForwarder:    badTxInterim,
		ArgsParser:        smartContract.NewArgumentParser(),
		ScrForwarder:      scForwarder,
		DisabledRelayedTx: false,
	}
	transactionProcessor, err := transaction.NewTxProcessor(argsNewTxProcessor)
	if err != nil {
		return nil, errors.New("could not create transaction statisticsProcessor: " + err.Error())
	}

	disabledRequestHandler := &disabled.RequestHandler{}
	disabledBlockTracker := &disabled.BlockTracker{}
	disabledBlockSizeComputationHandler := &disabled.BlockSizeComputationHandler{}
	disabledBalanceComputationHandler := &disabled.BalanceComputationHandler{}

	preProcFactory, err := shard.NewPreProcessorsContainerFactory(
		arg.ShardCoordinator,
		arg.Store,
		arg.Marshalizer,
		arg.Hasher,
		arg.DataPool,
		arg.PubkeyConv,
		arg.Accounts,
		disabledRequestHandler,
		transactionProcessor,
		scProcessor,
		scProcessor,
		rewardsTxProcessor,
		arg.Economics,
		gasHandler,
		disabledBlockTracker,
		disabledBlockSizeComputationHandler,
		disabledBalanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, err
	}

	txCoordinator, err := coordinator.NewTransactionCoordinator(
		arg.Hasher,
		arg.Marshalizer,
		arg.ShardCoordinator,
		arg.Accounts,
		arg.DataPool.MiniBlocks(),
		disabledRequestHandler,
		preProcContainer,
		interimProcContainer,
		gasHandler,
		genesisFeeHandler,
		disabledBlockSizeComputationHandler,
		disabledBalanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	queryService, err := smartContract.NewSCQueryService(vmContainer, arg.Economics)
	if err != nil {
		return nil, err
	}

	return &genesisProcessors{
		txCoordinator:  txCoordinator,
		systemSCs:      nil,
		txProcessor:    transactionProcessor,
		scProcessor:    scProcessor,
		scrProcessor:   scProcessor,
		rwdProcessor:   rewardsTxProcessor,
		blockchainHook: vmFactoryImpl.BlockChainHookImpl(),
		queryService:   queryService,
		vmContainer:    vmContainer,
	}, nil
}

func deployInitialSmartContracts(
	processors *genesisProcessors,
	arg ArgsGenesisBlockCreator,
	deployMetrics *deployedScMetrics,
) error {
	smartContracts, err := arg.SmartContractParser.InitialSmartContractsSplitOnOwnersShards(arg.ShardCoordinator)
	if err != nil {
		return err
	}

	currentShardSmartContracts := smartContracts[arg.ShardCoordinator.SelfId()]
	for _, sc := range currentShardSmartContracts {
		err = deployInitialSmartContract(processors, sc, arg, deployMetrics)
		if err != nil {
			return fmt.Errorf("%w for owner %s and filename %s",
				err, sc.GetOwner(), sc.GetFilename())
		}
	}

	return nil
}

func deployInitialSmartContract(
	processors *genesisProcessors,
	sc genesis.InitialSmartContractHandler,
	arg ArgsGenesisBlockCreator,
	deployMetrics *deployedScMetrics,
) error {

	txExecutor, err := intermediate.NewTxExecutionProcessor(processors.txProcessor, arg.Accounts)
	if err != nil {
		return err
	}

	var deployProc genesis.DeployProcessor

	switch sc.GetType() {
	case genesis.DNSType:
		deployMetrics.numOtherTypes++
		argDeployLibrary := intermediate.ArgDeployLibrarySC{
			Executor:         txExecutor,
			PubkeyConv:       arg.PubkeyConv,
			BlockchainHook:   processors.blockchainHook,
			ShardCoordinator: arg.ShardCoordinator,
		}
		deployProc, err = intermediate.NewDeployLibrarySC(argDeployLibrary)
		if err != nil {
			return err
		}
	case genesis.DelegationType:
		deployMetrics.numDelegation++
		fallthrough
	default:
		argDeploy := intermediate.ArgDeployProcessor{
			Executor:       txExecutor,
			PubkeyConv:     arg.PubkeyConv,
			BlockchainHook: processors.blockchainHook,
			QueryService:   processors.queryService,
		}
		deployProc, err = intermediate.NewDeployProcessor(argDeploy)
		if err != nil {
			return err
		}
	}

	return deployProc.Deploy(sc)
}

func increaseStakersNonces(processors *genesisProcessors, arg ArgsGenesisBlockCreator) (int, error) {
	txExecutor, err := intermediate.NewTxExecutionProcessor(processors.txProcessor, arg.Accounts)
	if err != nil {
		return 0, err
	}

	initialAddresses, err := arg.AccountsParser.InitialAccountsSplitOnAddressesShards(arg.ShardCoordinator)
	if err != nil {
		return 0, err
	}

	stakersCounter := 0
	initalAddressesInCurrentShard := initialAddresses[arg.ShardCoordinator.SelfId()]
	for _, ia := range initalAddressesInCurrentShard {
		if ia.GetStakingValue().Cmp(zero) < 1 {
			continue
		}

		numNodesStaked := big.NewInt(0).Set(ia.GetStakingValue())
		numNodesStaked.Div(numNodesStaked, arg.GenesisNodePrice)

		stakersCounter++
		err = txExecutor.AddNonce(ia.AddressBytes(), numNodesStaked.Uint64())
		if err != nil {
			return 0, fmt.Errorf("%w when adding nonce for address %s", err, ia.GetAddress())
		}
	}

	return stakersCounter, nil
}

func executeDelegation(
	processors *genesisProcessors,
	arg ArgsGenesisBlockCreator,
	nodesListSplitter genesis.NodesListSplitter,
) (genesis.DelegationResult, error) {
	txExecutor, err := intermediate.NewTxExecutionProcessor(processors.txProcessor, arg.Accounts)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	argDP := intermediate.ArgStandardDelegationProcessor{
		Executor:            txExecutor,
		ShardCoordinator:    arg.ShardCoordinator,
		AccountsParser:      arg.AccountsParser,
		SmartContractParser: arg.SmartContractParser,
		NodesListSplitter:   nodesListSplitter,
		QueryService:        processors.queryService,
		NodePrice:           arg.GenesisNodePrice,
	}

	delegationProcessor, err := intermediate.NewStandardDelegationProcessor(argDP)
	if err != nil {
		return genesis.DelegationResult{}, err
	}

	return delegationProcessor.ExecuteDelegation()
}

func incrementNoncesForCrossShardDelegations(processors *genesisProcessors, arg ArgsGenesisBlockCreator) (int, error) {
	txExecutor, err := intermediate.NewTxExecutionProcessor(processors.txProcessor, arg.Accounts)
	if err != nil {
		return 0, err
	}

	initialAddresses, err := arg.AccountsParser.InitialAccountsSplitOnAddressesShards(arg.ShardCoordinator)
	if err != nil {
		return 0, err
	}

	counter := 0
	initalAddressesInCurrentShard := initialAddresses[arg.ShardCoordinator.SelfId()]
	for _, ia := range initalAddressesInCurrentShard {
		dh := ia.GetDelegationHandler()
		if check.IfNil(dh) {
			continue
		}
		if arg.ShardCoordinator.SameShard(ia.AddressBytes(), dh.AddressBytes()) {
			continue
		}

		counter++
		err = txExecutor.AddNonce(ia.AddressBytes(), 1)
		if err != nil {
			return 0, fmt.Errorf("%w when adding nonce for address %s", err, ia.GetAddress())
		}
	}

	return counter, nil
}
