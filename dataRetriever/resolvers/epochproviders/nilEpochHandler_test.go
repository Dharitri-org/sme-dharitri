package epochproviders

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/stretchr/testify/require"
)

func TestNilEpochHandler_Epoch(t *testing.T) {
	t.Parallel()

	nilEpoch := &nilEpochHandler{}

	require.False(t, check.IfNil(nilEpoch))
	require.Equal(t, uint32(0), nilEpoch.MetaEpoch())
}
