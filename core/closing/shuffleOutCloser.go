package closing

import (
	"context"
	"fmt"
	"time"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/random"
	"github.com/Dharitri-org/sme-dharitri/data/endProcess"
	logger "github.com/Dharitri-org/sme-logger"
)

const minDuration = time.Second

var log = logger.GetOrCreate("core/closing")

type shuffleOutCloser struct {
	minWaitDuration time.Duration
	maxWaitDuration time.Duration
	signalChan      chan endProcess.ArgEndProcess
	randomizer      IntRandomizer
	ctx             context.Context
	cancelFunc      func()
}

// NewShuffleOutCloser creates a shuffle out component that is able to trigger a node restart and cancel that request if necessarily
func NewShuffleOutCloser(
	minWaitDuration time.Duration,
	maxWaitDuration time.Duration,
	signalChan chan endProcess.ArgEndProcess,
) (*shuffleOutCloser, error) {

	if minWaitDuration < minDuration {
		return nil, fmt.Errorf("%w for minWaitDuration", core.ErrInvalidValue)
	}
	if maxWaitDuration < minDuration {
		return nil, fmt.Errorf("%w for maxWaitDuration", core.ErrInvalidValue)
	}
	if minWaitDuration > maxWaitDuration {
		return nil, fmt.Errorf("%w, minWaitDuration > maxWaitDuration", core.ErrInvalidValue)
	}
	if signalChan == nil {
		return nil, core.ErrNilSignalChan
	}

	soc := &shuffleOutCloser{
		minWaitDuration: minWaitDuration,
		maxWaitDuration: maxWaitDuration,
		signalChan:      signalChan,
		randomizer:      &random.ConcurrentSafeIntRandomizer{},
	}
	soc.ctx, soc.cancelFunc = context.WithCancel(context.Background())

	return soc, nil
}

// EndOfProcessingHandler will be called each time a delayed end of processing is needed
func (soc *shuffleOutCloser) EndOfProcessingHandler(event endProcess.ArgEndProcess) error {
	go soc.writeOnChanDelayed(event)

	return nil
}

func (soc *shuffleOutCloser) writeOnChanDelayed(event endProcess.ArgEndProcess) {
	delta := soc.maxWaitDuration - soc.minWaitDuration

	randDurationBeforeStop := soc.randomizer.Intn(int(delta))
	timeToWait := soc.minWaitDuration + time.Duration(randDurationBeforeStop)

	log.Info("the application will stop in",
		"waiting time", fmt.Sprintf("%v", timeToWait),
		"description", event.Description,
		"reason", event.Reason)

	select {
	case <-time.After(timeToWait):
	case <-soc.ctx.Done():
		log.Debug("canceled the application stop go routine")
		return
	}

	log.Info("the application will stop now after",
		"waiting time", fmt.Sprintf("%v", timeToWait),
		"description", event.Description,
		"reason", event.Reason,
	)

	soc.signalChan <- event
}

// Close cancels the channel write
func (soc *shuffleOutCloser) Close() error {
	soc.cancelFunc()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (soc *shuffleOutCloser) IsInterfaceNil() bool {
	return soc == nil
}
