package disabled

import (
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/process"
)

var _ process.InterceptedHeaderSigVerifier = (*headerSigVerifier)(nil)

type headerSigVerifier struct {
}

// NewHeaderSigVerifier returns a new instance of headerSigVerifier
func NewHeaderSigVerifier() *headerSigVerifier {
	return &headerSigVerifier{}
}

// VerifyRandSeedAndLeaderSignature -
func (h *headerSigVerifier) VerifyRandSeedAndLeaderSignature(_ data.HeaderHandler) error {
	return nil
}

// VerifySignature -
func (h *headerSigVerifier) VerifySignature(_ data.HeaderHandler) error {
	return nil
}

// IsInterfaceNil -
func (h *headerSigVerifier) IsInterfaceNil() bool {
	return h == nil
}
