package mock

import (
	"github.com/Dharitri-org/sme-dharitri/data"
)

// BlockNotifierHandlerMock -
type BlockNotifierHandlerMock struct {
	CallHandlersCalled    func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	RegisterHandlerCalled func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
}

// CallHandlers -
func (bnhm *BlockNotifierHandlerMock) CallHandlers(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	if bnhm.CallHandlersCalled != nil {
		bnhm.CallHandlersCalled(shardID, headers, headersHashes)
	}
}

// RegisterHandler -
func (bnhm *BlockNotifierHandlerMock) RegisterHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bnhm.RegisterHandlerCalled != nil {
		bnhm.RegisterHandlerCalled(handler)
	}
}

// IsInterfaceNil -
func (bnhm *BlockNotifierHandlerMock) IsInterfaceNil() bool {
	return bnhm == nil
}
