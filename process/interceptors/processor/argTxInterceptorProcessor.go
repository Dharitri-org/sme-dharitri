package processor

import (
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
	"github.com/Dharitri-org/sme-dharitri/process"
)

// ArgTxInterceptorProcessor is the argument for the interceptor processor used for transactions
// (balance txs, smart contract results, reward and so on)
type ArgTxInterceptorProcessor struct {
	ShardedDataCache dataRetriever.ShardedDataCacherNotifier
	TxValidator      process.TxValidator
}
