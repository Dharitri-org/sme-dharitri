package transaction

import (
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/process"
)

type TxProcessor *txProcessor

func (txProc *txProcessor) GetAccounts(adrSrc, adrDst []byte,
) (acntSrc, acntDst state.UserAccountHandler, err error) {
	return txProc.getAccounts(adrSrc, adrDst)
}

func (txProc *txProcessor) CheckTxValues(tx *transaction.Transaction, acntSnd, acntDst state.UserAccountHandler) error {
	return txProc.checkTxValues(tx, acntSnd, acntDst)
}

func (txProc *txProcessor) IncreaseNonce(acntSrc state.UserAccountHandler) {
	acntSrc.IncreaseNonce(1)
}

func (txProc *txProcessor) ProcessTxFee(
	tx *transaction.Transaction,
	acntSnd, acntDst state.UserAccountHandler,
) (*big.Int, error) {
	return txProc.processTxFee(tx, acntSnd, acntDst)
}

func (inTx *InterceptedTransaction) SetWhitelistHandler(handler process.WhiteListHandler) {
	inTx.whiteListerVerifiedTxs = handler
}
