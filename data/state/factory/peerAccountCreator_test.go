package factory_test

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/state/factory"
	"github.com/stretchr/testify/assert"
)

func TestPeerAccountCreator_CreateAccountNilAddress(t *testing.T) {
	t.Parallel()

	accF := factory.NewPeerAccountCreator()

	_, ok := accF.(*factory.PeerAccountCreator)
	assert.Equal(t, true, ok)

	acc, err := accF.CreateAccount(nil)

	assert.Nil(t, acc)
	assert.Equal(t, err, state.ErrNilAddress)
}

func TestPeerAccountCreator_CreateAccountOk(t *testing.T) {
	t.Parallel()

	accF := factory.NewPeerAccountCreator()
	assert.False(t, check.IfNil(accF))

	_, ok := accF.(*factory.PeerAccountCreator)
	assert.Equal(t, true, ok)

	acc, err := accF.CreateAccount(make([]byte, 32))

	assert.NotNil(t, acc)
	assert.Nil(t, err)
}
