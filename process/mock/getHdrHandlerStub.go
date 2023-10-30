package mock

import "github.com/Dharitri-org/sme-dharitri/data"

// GetHdrHandlerStub -
type GetHdrHandlerStub struct {
	HeaderHandlerCalled func() data.HeaderHandler
}

// HeaderHandler -
func (ghhs *GetHdrHandlerStub) HeaderHandler() data.HeaderHandler {
	return ghhs.HeaderHandlerCalled()
}
