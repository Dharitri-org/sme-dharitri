package process

import (
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/trie"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
)

const maxTrieLevelInMemory = uint(5)

func createAccountAdapter(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	accountFactory state.AccountFactory,
	trieStorage data.StorageManager,
) (state.AccountsAdapter, error) {
	tr, err := trie.NewTrie(trieStorage, marshalizer, hasher, maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	adb, err := state.NewAccountsDB(tr, hasher, marshalizer, accountFactory)
	if err != nil {
		return nil, err
	}

	return adb, nil
}
