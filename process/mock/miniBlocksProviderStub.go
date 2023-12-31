package mock

import (
	"github.com/Dharitri-org/sme-dharitri/process"
)

// MiniBlocksProviderStub -
type MiniBlocksProviderStub struct {
	GetMiniBlocksCalled         func(hashes [][]byte) ([]*process.MiniblockAndHash, [][]byte)
	GetMiniBlocksFromPoolCalled func(hashes [][]byte) ([]*process.MiniblockAndHash, [][]byte)
}

// GetMiniBlocks -
func (mbps *MiniBlocksProviderStub) GetMiniBlocks(hashes [][]byte) ([]*process.MiniblockAndHash, [][]byte) {
	if mbps.GetMiniBlocksCalled != nil {
		return mbps.GetMiniBlocksCalled(hashes)
	}
	return nil, nil
}

// GetMiniBlocksFromPool -
func (mbps *MiniBlocksProviderStub) GetMiniBlocksFromPool(hashes [][]byte) ([]*process.MiniblockAndHash, [][]byte) {
	if mbps.GetMiniBlocksFromPoolCalled != nil {
		return mbps.GetMiniBlocksFromPoolCalled(hashes)
	}
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mbps *MiniBlocksProviderStub) IsInterfaceNil() bool {
	return mbps == nil
}
