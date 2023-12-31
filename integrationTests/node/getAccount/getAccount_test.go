package getAccount

import (
	"math/big"
	"testing"

	"github.com/Dharitri-org/sme-dharitri/integrationTests"
	"github.com/Dharitri-org/sme-dharitri/node"
	"github.com/stretchr/testify/assert"
)

func TestNode_GetAccountAccountDoesNotExistsShouldRetEmpty(t *testing.T) {
	t.Parallel()

	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	accDB, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressPubkeyConverter(integrationTests.TestAddressPubkeyConverter),
	)

	encodedAddress := integrationTests.TestAddressPubkeyConverter.Encode(integrationTests.CreateRandomBytes(32))
	recovAccnt, err := n.GetAccount(encodedAddress)

	assert.Nil(t, err)
	assert.Equal(t, uint64(0), recovAccnt.GetNonce())
	assert.Equal(t, big.NewInt(0), recovAccnt.GetBalance())
	assert.Nil(t, recovAccnt.GetCodeHash())
	assert.Nil(t, recovAccnt.GetRootHash())
}

func TestNode_GetAccountAccountExistsShouldReturn(t *testing.T) {
	t.Parallel()

	trieStorage, _ := integrationTests.CreateTrieStorageManager()
	accDB, _ := integrationTests.CreateAccountsDB(0, trieStorage)

	addressBytes := integrationTests.CreateRandomBytes(32)
	nonce := uint64(2233)
	account, _ := accDB.LoadAccount(addressBytes)
	account.IncreaseNonce(nonce)
	_ = accDB.SaveAccount(account)
	_, _ = accDB.Commit()

	n, _ := node.NewNode(
		node.WithAccountsAdapter(accDB),
		node.WithAddressPubkeyConverter(integrationTests.TestAddressPubkeyConverter),
	)

	encodedAddress := integrationTests.TestAddressPubkeyConverter.Encode(addressBytes)
	recovAccnt, err := n.GetAccount(encodedAddress)

	assert.Nil(t, err)
	assert.Equal(t, nonce, recovAccnt.GetNonce())
}
