package config

// SystemSmartContractsConfig defines the system smart contract configs
type SystemSmartContractsConfig struct {
	DCTSystemSCConfig        DCTSystemSCConfig
	GovernanceSystemSCConfig GovernanceSystemSCConfig
}

// DCTSystemSCConfig defines a set of constant to initialize the dct system smart contract
type DCTSystemSCConfig struct {
	BaseIssuingCost string
	OwnerAddress    string
}

// GovernanceSystemSCConfig defines the set of constants to initialize the governance system smart contract
type GovernanceSystemSCConfig struct {
	ProposalCost     string
	NumNodes         int64
	MinQuorum        int32
	MinPassThreshold int32
	MinVetoThreshold int32
}
