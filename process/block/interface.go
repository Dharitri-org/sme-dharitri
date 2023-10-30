package block

import (
	"github.com/Dharitri-org/sme-dharitri/data"
)

type blockProcessor interface {
	removeStartOfEpochBlockDataFromPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error
}
