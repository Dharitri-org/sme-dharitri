package dataRetriever

import (
	"github.com/Dharitri-org/sme-dharitri/process/factory"
)

// SetEpochHandlerToHdrResolver sets the epoch handler to the metablock hdr resolver
func SetEpochHandlerToHdrResolver(
	resolversContainer ResolversContainer,
	epochHandler EpochHandler,
) error {
	resolver, err := resolversContainer.Get(factory.MetachainBlocksTopic)
	if err != nil {
		return err
	}

	hdrResolver, ok := resolver.(HeaderResolver)
	if !ok {
		return ErrWrongTypeInContainer
	}

	err = hdrResolver.SetEpochHandler(epochHandler)
	if err != nil {
		return err
	}

	return nil
}
