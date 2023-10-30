package notifier

import (
	"sync"

	"github.com/Dharitri-org/sme-dharitri/epochStart"
)

func (essh *epochStartSubscriptionHandler) RegisteredHandlers() ([]epochStart.ActionHandler, *sync.RWMutex) {
	return essh.epochStartHandlers, &essh.mutEpochStartHandler
}
