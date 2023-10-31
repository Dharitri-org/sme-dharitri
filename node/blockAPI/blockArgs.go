package blockAPI

import (
	"github.com/Dharitri-org/sme-dharitri/core/fullHistory"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/data/typeConverters"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/marshal"
)

// APIBlockProcessorArg is structure that store components that are needed to create an api block procesosr
type APIBlockProcessorArg struct {
	SelfShardID              uint32
	Store                    dataRetriever.StorageService
	Marshalizer              marshal.Marshalizer
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	HistoryRepo              fullHistory.HistoryRepository
	UnmarshalTx              func(txBytes []byte, txType string) (*transaction.ApiTransactionResult, error)
}
