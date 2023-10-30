package mock

import "github.com/Dharitri-org/sme-dharitri/data"

// HeaderSigVerifierStub -
type HeaderSigVerifierStub struct {
	VerifyRandSeedCaller func(header data.HeaderHandler) error
}

// VerifyRandSeed -
func (hsvm *HeaderSigVerifierStub) VerifyRandSeed(header data.HeaderHandler) error {
	if hsvm.VerifyRandSeedCaller != nil {
		return hsvm.VerifyRandSeedCaller(header)
	}

	return nil
}

// IsInterfaceNil -
func (hsvm *HeaderSigVerifierStub) IsInterfaceNil() bool {
	return hsvm == nil
}
