package rewardTransaction

import (
	"github.com/Dharitri-org/sme-dharitri/hashing"
)

// RewardKey -
const RewardKey = rewardKey

// Hasher will return the hasher of InterceptedRewardTransaction for using in test files
func (inRTx *InterceptedRewardTransaction) Hasher() hashing.Hasher {
	return inRTx.hasher
}
