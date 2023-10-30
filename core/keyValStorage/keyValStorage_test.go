package keyValStorage_test

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core/keyValStorage"
	"github.com/stretchr/testify/assert"
)

func TestNewKeyValStorage_GetKeyAndVal(t *testing.T) {
	t.Parallel()

	key := []byte("key")
	value := []byte("value")

	keyVal := keyValStorage.NewKeyValStorage(key, value)
	assert.NotNil(t, keyVal)
	assert.Equal(t, key, keyVal.GetKey())
	assert.Equal(t, value, keyVal.GetValue())
}