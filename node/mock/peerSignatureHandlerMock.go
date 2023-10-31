package mock

import (
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/crypto"
)

// PeerSignatureHandler -
type PeerSignatureHandler struct{}

// VerifyPeerSignature -
func (p *PeerSignatureHandler) VerifyPeerSignature(_ []byte, _ core.PeerID, _ []byte) error {
	return nil
}

// GetPeerSignature -
func (p *PeerSignatureHandler) GetPeerSignature(_ crypto.PrivateKey, _ []byte) ([]byte, error) {
	return nil, nil
}

// IsInterfaceNil -
func (p *PeerSignatureHandler) IsInterfaceNil() bool {
	return p == nil
}
