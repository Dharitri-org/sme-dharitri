package mock

import (
	"context"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/process"
)

// MetaBlockInterceptorProcessorStub -
type MetaBlockInterceptorProcessorStub struct {
	GetEpochStartMetaBlockCalled func() (*block.MetaBlock, error)
}

// Validate -
func (m *MetaBlockInterceptorProcessorStub) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save -
func (m *MetaBlockInterceptorProcessorStub) Save(_ process.InterceptedData, _ core.PeerID, _ string) error {
	return nil
}

// RegisterHandler -
func (m *MetaBlockInterceptorProcessorStub) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
}

// SignalEndOfProcessing -
func (m *MetaBlockInterceptorProcessorStub) SignalEndOfProcessing(_ []process.InterceptedData) {
}

// IsInterfaceNil -
func (m *MetaBlockInterceptorProcessorStub) IsInterfaceNil() bool {
	return m == nil
}

// GetEpochStartMetaBlock -
func (m *MetaBlockInterceptorProcessorStub) GetEpochStartMetaBlock(_ context.Context) (*block.MetaBlock, error) {
	if m.GetEpochStartMetaBlockCalled != nil {
		return m.GetEpochStartMetaBlockCalled()
	}

	return &block.MetaBlock{}, nil
}
