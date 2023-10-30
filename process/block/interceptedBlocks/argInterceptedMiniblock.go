package interceptedBlocks

import (
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// ArgInterceptedMiniblock is the argument for the intercepted miniblock
type ArgInterceptedMiniblock struct {
	MiniblockBuff    []byte
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
}
