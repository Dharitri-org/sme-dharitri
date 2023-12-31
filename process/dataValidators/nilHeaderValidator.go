package dataValidators

import (
	"github.com/Dharitri-org/sme-dharitri/process"
)

var _ process.HeaderValidator = (*nilHeaderValidator)(nil)

// nilHeaderValidator represents a header handler validator that doesn't check the validity of provided headerHandler
type nilHeaderValidator struct {
}

// NewNilHeaderValidator creates a new nil header handler validator instance
func NewNilHeaderValidator() (*nilHeaderValidator, error) {
	return &nilHeaderValidator{}, nil
}

// HeaderValidForProcessing is a nil implementation that will return true
func (nhv *nilHeaderValidator) HeaderValidForProcessing(process.HdrValidatorHandler) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nhv *nilHeaderValidator) IsInterfaceNil() bool {
	return nhv == nil
}
