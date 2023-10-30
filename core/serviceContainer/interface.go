package serviceContainer

import (
	"github.com/Dharitri-org/sme-dharitri/core/indexer"
	"github.com/Dharitri-org/sme-dharitri/core/statistics"
)

// Core interface will abstract all the subpackage functionalities and will
//
//	provide access to it's members where needed
type Core interface {
	Indexer() indexer.Indexer
	TPSBenchmark() statistics.TPSBenchmark
	IsInterfaceNil() bool
}
