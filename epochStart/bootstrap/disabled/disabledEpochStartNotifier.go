package disabled

import (
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/epochStart"
)

// EpochStartNotifier -
type EpochStartNotifier struct {
}

// RegisterHandler -
func (desn *EpochStartNotifier) RegisterHandler(_ epochStart.ActionHandler) {
}

// UnregisterHandler -
func (desn *EpochStartNotifier) UnregisterHandler(_ epochStart.ActionHandler) {
}

// NotifyAllPrepare -
func (desn *EpochStartNotifier) NotifyAllPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NotifyAll -
func (desn *EpochStartNotifier) NotifyAll(_ data.HeaderHandler) {
}

// IsInterfaceNil -
func (desn *EpochStartNotifier) IsInterfaceNil() bool {
	return desn == nil
}
