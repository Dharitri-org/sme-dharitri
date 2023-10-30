package processor

import (
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	"github.com/Dharitri-org/sme-dharitri/storage"
)

// ArgMiniblockInterceptorProcessor is the argument for the interceptor processor used for miniblocks
type ArgMiniblockInterceptorProcessor struct {
	MiniblockCache   storage.Cacher
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	ShardCoordinator sharding.Coordinator
	WhiteListHandler process.WhiteListHandler
}
