package factory

import (
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/storage"
)

// UserAccountTrie represents the use account identifier
const UserAccountTrie = "userAccount"

// PeerAccountTrie represents the peer account identifier
const PeerAccountTrie = "peerAccount"

// TrieFactoryArgs holds arguments for creating a trie factory
type TrieFactoryArgs struct {
	EvictionWaitingListCfg   config.EvictionWaitingListConfig
	SnapshotDbCfg            config.DBConfig
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	PathManager              storage.PathManagerHandler
	TrieStorageManagerConfig config.TrieStorageManagerConfig
}
