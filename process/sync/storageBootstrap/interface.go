package storageBootstrap

import (
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/process/block/bootstrapStorage"
)

// StorageBootstrapper is the main interface for bootstrap from storage execution engine
type storageBootstrapperHandler interface {
	getHeader(hash []byte) (data.HeaderHandler, error)
	getHeaderWithNonce(nonce uint64, shardID uint32) (data.HeaderHandler, []byte, error)
	applyCrossNotarizedHeaders(crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) error
	applyNumPendingMiniBlocks(pendingMiniBlocks []bootstrapStorage.PendingMiniBlocksInfo)
	applySelfNotarizedHeaders(selfNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) ([]data.HeaderHandler, [][]byte, error)
	cleanupNotarizedStorage(hash []byte)
	IsInterfaceNil() bool
}
