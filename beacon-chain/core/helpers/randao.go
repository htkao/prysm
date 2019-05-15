package helpers

import (
	"fmt"

	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/hashutil"
	"github.com/prysmaticlabs/prysm/shared/params"
)

// GenerateSeed generates the randao seed of a given epoch.
//
// Spec pseudocode definition:
//   def generate_seed(state: BeaconState,
// 	            epoch: Epoch) -> Bytes32:
//    """
//    Generate a seed for the given ``epoch``.
//    """
//    return hash(
//        get_randao_mix(state, epoch - MIN_SEED_LOOKAHEAD) +
//        get_active_index_root(state, epoch) +
//        int_to_bytes32(epoch)
//    )
func GenerateSeed(state *pb.BeaconState, wantedEpoch uint64) ([32]byte, error) {
	randaoMix := RandaoMix(state, wantedEpoch-params.BeaconConfig().MinSeedLookahead)

	indexRoot, err := ActiveIndexRoot(state, wantedEpoch)
	if err != nil {
		return [32]byte{}, err
	}
	th := append(randaoMix, indexRoot...)
	th = append(th, bytesutil.Bytes32(wantedEpoch)...)
	return hashutil.Hash(th), nil
}

// ActiveIndexRoot returns the index root of a given epoch.
//
// Spec pseudocode definition:
//   def get_active_index_root(state: BeaconState,
//                          epoch: Epoch) -> Bytes32:
//    """
//    Return the index root at a recent ``epoch``.
//    """
//    assert get_current_epoch(state) - LATEST_ACTIVE_INDEX_ROOTS_LENGTH + ACTIVATION_EXIT_DELAY < epoch <= get_current_epoch(state) + ACTIVATION_EXIT_DELAY
//	  return state.latest_active_index_roots[epoch % LATEST_ACTIVE_INDEX_ROOTS_LENGTH]
func ActiveIndexRoot(state *pb.BeaconState, wantedEpoch uint64) ([]byte, error) {
	var earliestEpoch uint64
	currentEpoch := CurrentEpoch(state)
	if currentEpoch > params.BeaconConfig().LatestActiveIndexRootsLength+params.BeaconConfig().ActivationExitDelay {
		earliestEpoch = currentEpoch - (params.BeaconConfig().LatestActiveIndexRootsLength + params.BeaconConfig().ActivationExitDelay)
	}
	if earliestEpoch > wantedEpoch || wantedEpoch > currentEpoch+params.BeaconConfig().ActivationExitDelay {
		return nil, fmt.Errorf("input indexRoot epoch %d out of bounds: %d <= epoch < %d",
			wantedEpoch, earliestEpoch, currentEpoch+params.BeaconConfig().ActivationExitDelay)
	}
	return state.LatestActiveIndexRoots[wantedEpoch%params.BeaconConfig().LatestActiveIndexRootsLength], nil
}

// RandaoMix returns the randao mix (xor'ed seed)
// of a given slot. It is used to shuffle validators.
//
// Spec pseudocode definition:
//   def get_randao_mix(state: BeaconState,
//                   epoch: Epoch) -> Bytes32:
//    """
//    Return the randao mix at a recent ``epoch``.
//    ``epoch`` expected to be between (current_epoch - LATEST_RANDAO_MIXES_LENGTH, current_epoch].
//    """
//    return state.latest_randao_mixes[epoch % LATEST_RANDAO_MIXES_LENGTH]
func RandaoMix(state *pb.BeaconState, epoch uint64) []byte {
	return state.LatestRandaoMixes[epoch%params.BeaconConfig().LatestRandaoMixesLength]
}
