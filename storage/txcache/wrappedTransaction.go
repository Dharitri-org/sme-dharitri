package txcache

import (
	"bytes"

	"github.com/Dharitri-org/sme-dharitri/data"
)

// WrappedTransaction contains a transaction, its hash and extra information
type WrappedTransaction struct {
	Tx              data.TransactionHandler
	TxHash          []byte
	SenderShardID   uint32
	ReceiverShardID uint32
	Size            int64
}

func (wrappedTx *WrappedTransaction) sameAs(another *WrappedTransaction) bool {
	return bytes.Equal(wrappedTx.TxHash, another.TxHash)
}

// estimateTxGas returns an approximation for the necessary computation units (gas units)
func estimateTxGas(tx *WrappedTransaction) uint64 {
	gasLimit := tx.Tx.GetGasLimit()
	return gasLimit
}

// estimateTxFee returns an approximation for the cost of a transaction, in nano MOA
// TODO: switch to integer operations (as opposed to float operations).
// TODO: do not assume the order of magnitude of minGasPrice.
func estimateTxFee(tx *WrappedTransaction) uint64 {
	// In order to obtain the result as nano MOA (not as "atomic" 10^-18 MOA), we have to divide by 10^9
	// In order to have better precision, we divide the factors by 10^6, and 10^3 respectively
	gasLimit := float32(tx.Tx.GetGasLimit()) / 1000000
	gasPrice := float32(tx.Tx.GetGasPrice()) / 1000
	feeInNanoMOA := gasLimit * gasPrice
	return uint64(feeInNanoMOA)
}
