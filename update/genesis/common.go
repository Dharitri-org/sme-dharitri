package genesis

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/sharding"
)

// TODO: create a structure or use this function also in process/peer/process.go
func getValidatorDataFromLeaves(
	leaves map[string][]byte,
	shardCoordinator sharding.Coordinator,
	marshalizer marshal.Marshalizer,
) (map[uint32][]*state.ValidatorInfo, error) {

	validators := make(map[uint32][]*state.ValidatorInfo, shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		validators[i] = make([]*state.ValidatorInfo, 0)
	}
	validators[core.MetachainShardId] = make([]*state.ValidatorInfo, 0)

	sliceLeaves := convertMapToSortedSlice(leaves)

	sort.Slice(sliceLeaves, func(i, j int) bool {
		return bytes.Compare(sliceLeaves[i], sliceLeaves[j]) < 0
	})

	for _, pa := range sliceLeaves {
		peerAccount, err := unmarshalPeer(pa, marshalizer)
		if err != nil {
			return nil, err
		}

		currentShardId := peerAccount.GetShardId()
		validatorInfoData := peerAccountToValidatorInfo(peerAccount)
		validators[currentShardId] = append(validators[currentShardId], validatorInfoData)
	}

	return validators, nil
}

func convertMapToSortedSlice(leaves map[string][]byte) [][]byte {
	newLeaves := make([][]byte, len(leaves))
	i := 0
	for _, pa := range leaves {
		newLeaves[i] = pa
		i++
	}

	return newLeaves
}

func unmarshalPeer(pa []byte, marshalizer marshal.Marshalizer) (state.PeerAccountHandler, error) {
	peerAccount := state.NewEmptyPeerAccount()
	err := marshalizer.Unmarshal(peerAccount, pa)
	if err != nil {
		return nil, err
	}
	return peerAccount, nil
}

func peerAccountToValidatorInfo(peerAccount state.PeerAccountHandler) *state.ValidatorInfo {
	return &state.ValidatorInfo{
		PublicKey:                  peerAccount.GetBLSPublicKey(),
		ShardId:                    peerAccount.GetShardId(),
		List:                       getActualList(peerAccount),
		Index:                      peerAccount.GetIndexInList(),
		TempRating:                 peerAccount.GetTempRating(),
		Rating:                     peerAccount.GetRating(),
		RewardAddress:              peerAccount.GetRewardAddress(),
		LeaderSuccess:              peerAccount.GetLeaderSuccessRate().NumSuccess,
		LeaderFailure:              peerAccount.GetLeaderSuccessRate().NumFailure,
		ValidatorSuccess:           peerAccount.GetValidatorSuccessRate().NumSuccess,
		ValidatorFailure:           peerAccount.GetValidatorSuccessRate().NumFailure,
		TotalLeaderSuccess:         peerAccount.GetTotalLeaderSuccessRate().NumSuccess,
		TotalLeaderFailure:         peerAccount.GetTotalLeaderSuccessRate().NumFailure,
		TotalValidatorSuccess:      peerAccount.GetTotalValidatorSuccessRate().NumSuccess,
		TotalValidatorFailure:      peerAccount.GetTotalValidatorSuccessRate().NumFailure,
		NumSelectedInSuccessBlocks: peerAccount.GetNumSelectedInSuccessBlocks(),
		AccumulatedFees:            big.NewInt(0).Set(peerAccount.GetAccumulatedFees()),
	}
}

func getActualList(peerAccount state.PeerAccountHandler) string {
	savedList := peerAccount.GetList()
	if peerAccount.GetUnStakedEpoch() == core.DefaultUnstakedEpoch {
		if savedList == string(core.InactiveList) {
			return string(core.JailedList)
		}
		return savedList
	}
	if savedList == string(core.InactiveList) {
		return savedList
	}

	return string(core.LeavingList)
}

func shouldExportValidator(validator *state.ValidatorInfo, allowedLists []core.PeerType) bool {
	validatorList := validator.GetList()

	for _, list := range allowedLists {
		if validatorList == string(list) {
			return true
		}
	}

	return false
}
