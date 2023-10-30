package factory_test

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/state/factory"
	"github.com/stretchr/testify/assert"
)

func TestAccountCreator_CreateAccountNilAddress(t *testing.T) {
	t.Parallel()

	accF := factory.NewAccountCreator()

	_, ok := accF.(*factory.AccountCreator)
	assert.Equal(t, true, ok)
	assert.False(t, check.IfNil(accF))

	acc, err := accF.CreateAccount(nil)

	assert.Nil(t, acc)
	assert.Equal(t, err, state.ErrNilAddress)
}

func TestAccountCreator_CreateAccountOk(t *testing.T) {
	t.Parallel()

	accF := factory.NewAccountCreator()

	_, ok := accF.(*factory.AccountCreator)
	assert.Equal(t, true, ok)

	acc, err := accF.CreateAccount(make([]byte, 32))

	assert.Nil(t, err)
	assert.False(t, check.IfNil(acc))
}
