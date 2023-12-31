package txcache

import (
	"math"
)

var _ scoreComputer = (*defaultScoreComputer)(nil)

// TODO: the score formula should not be sensitive to the order of magnitude of the minGasPrice.
// TODO (continued): We should not rely on any order of magnitude known a priori.
// TODO (continued): The score formula should work even if minGasPrice = 0.
type senderScoreParams struct {
	count uint64
	// Size is in bytes
	size uint64
	// Fee is in nano MOA
	fee uint64
	gas uint64
}

type defaultScoreComputer struct {
	// Price is in nano MOA
	minGasPrice uint32
}

func newDefaultScoreComputer(minGasPrice uint32) *defaultScoreComputer {
	return &defaultScoreComputer{
		minGasPrice: minGasPrice,
	}
}

// computeScore computes the score of the sender, as an integer 0-100
func (computer *defaultScoreComputer) computeScore(scoreParams senderScoreParams) uint32 {
	rawScore := computer.computeRawScore(scoreParams)
	truncatedScore := uint32(rawScore)
	return truncatedScore
}

// TODO (optimization): switch to integer operations (as opposed to float operations).
func (computer *defaultScoreComputer) computeRawScore(params senderScoreParams) float64 {
	allParamsDefined := params.fee > 0 && params.gas > 0 && params.size > 0 && params.count > 0
	if !allParamsDefined {
		return 0
	}

	PPUMin := float64(computer.minGasPrice)
	PPUAvg := float64(params.fee) / float64(params.gas)
	PPUScore := math.Pow(PPUAvg/PPUMin, 3)

	countPow2 := float64(params.count) * float64(params.count)
	countScore := math.Log(countPow2+1) + 1

	// We use size in ~kB
	const bytesInKB = 1000
	size := float64(params.size) / bytesInKB
	sizePow2 := size * size
	sizeScore := math.Log(sizePow2+1) + 1

	rawScore := PPUScore / countScore / sizeScore
	// We apply the logistic function,
	// and then subtract 0.5, since we only deal with positive scores,
	// and then we multiply by 2, to have full [0..1] range.
	asymptoticScore := (1/(1+math.Exp(-rawScore)) - 0.5) * 2
	score := asymptoticScore * float64(numberOfScoreChunks)
	return score
}
