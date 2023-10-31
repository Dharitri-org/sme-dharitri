package process

import (
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/typeConverters"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/genesis"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/economics"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/update"
)

// ArgsGenesisBlockCreator holds the arguments which are needed to create a genesis block
type ArgsGenesisBlockCreator struct {
	GenesisTime              uint64
	StartEpochNum            uint32
	Accounts                 state.AccountsAdapter
	ValidatorAccounts        state.AccountsAdapter
	PubkeyConv               core.PubkeyConverter
	InitialNodesSetup        genesis.InitialNodesHandler
	Economics                *economics.EconomicsData //TODO refactor and use an interface
	ShardCoordinator         sharding.Coordinator
	Store                    dataRetriever.StorageService
	Blkc                     data.ChainHandler
	Marshalizer              marshal.Marshalizer
	SignMarshalizer          marshal.Marshalizer
	Hasher                   hashing.Hasher
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	DataPool                 dataRetriever.PoolsHolder
	AccountsParser           genesis.AccountsParser
	SmartContractParser      genesis.InitialSmartContractParser
	GasMap                   map[string]map[string]uint64
	TxLogsProcessor          process.TransactionLogProcessor
	VirtualMachineConfig     config.VirtualMachineConfig
	HardForkConfig           config.HardforkConfig
	TrieStorageManagers      map[string]data.StorageManager
	ChainID                  string
	SystemSCConfig           config.SystemSmartContractsConfig
	BlockSignKeyGen          crypto.KeyGenerator
	ImportStartHandler       update.ImportStartHandler
	WorkingDir               string
	GenesisNodePrice         *big.Int
	GenesisString            string
	// created components
	importHandler update.ImportHandler
}
