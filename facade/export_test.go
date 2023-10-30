package facade

import (
	"github.com/Dharitri-org/sme-dharitri/ntp"
)

// GetSyncer returns the current syncer
func (nf *nodeFacade) GetSyncer() ntp.SyncTimer {
	return nf.syncer
}
