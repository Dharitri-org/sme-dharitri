package disabled

import (
	"github.com/Dharitri-org/sme-dharitri/process"
)

type disabledWhiteListVerifier struct {
}

// NewDisabledWhiteListDataVerifier returns a default data verifier
func NewDisabledWhiteListDataVerifier() (*disabledWhiteListVerifier, error) {
	return &disabledWhiteListVerifier{}, nil
}

// IsWhiteListed return true if intercepted data is accepted
func (w *disabledWhiteListVerifier) IsWhiteListed(_ process.InterceptedData) bool {
	return true
}

// Add adds all the list to the cache
func (w *disabledWhiteListVerifier) Add(_ [][]byte) {
}

// Remove removes all the keys from the cache
func (w *disabledWhiteListVerifier) Remove(_ [][]byte) {
}

// IsInterfaceNil returns true if underlying object is nil
func (w *disabledWhiteListVerifier) IsInterfaceNil() bool {
	return w == nil
}
