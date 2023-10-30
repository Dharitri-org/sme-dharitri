package peer

import (
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
)

// DataPool indicates the main functionality needed in order to fetch the required blocks from the pool
type DataPool interface {
	Headers() dataRetriever.HeadersPool
	IsInterfaceNil() bool
}
