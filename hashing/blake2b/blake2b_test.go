package blake2b_test

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/hashing/blake2b"
	"github.com/stretchr/testify/assert"
)

func TestBlake2b_ComputeWithDifferentHashSizes(t *testing.T) {
	t.Parallel()

	input := "dummy string"
	sizes := []int{2, 5, 8, 16, 32, 37, 64}
	for _, size := range sizes {
		testComputeOk(t, input, size)
	}
}

func testComputeOk(t *testing.T, input string, size int) {
	hasher := blake2b.Blake2b{HashSize: size}
	res := hasher.Compute(input)
	assert.Equal(t, size, len(res))
}

func TestBlake2b_Empty(t *testing.T) {

	hasher := &blake2b.Blake2b{HashSize: 64}

	var nilStr string
	res_nil := hasher.Compute(nilStr)
	assert.Equal(t, 64, len(res_nil))

	res_empty := hasher.Compute("")
	assert.Equal(t, 64, len(res_empty))

	assert.Equal(t, res_empty, res_nil)

	// force recompute
	hasher = &blake2b.Blake2b{HashSize: 64}

	res_empty = hasher.Compute("")
	assert.Equal(t, 64, len(res_empty))

	assert.Equal(t, res_empty, res_nil)
}
