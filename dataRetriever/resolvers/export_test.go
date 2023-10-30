package resolvers

import "github.com/Dharitri-org/sme-dharitri/dataRetriever"

func (hdrRes *HeaderResolver) EpochHandler() dataRetriever.EpochHandler {
	return hdrRes.epochHandler
}
