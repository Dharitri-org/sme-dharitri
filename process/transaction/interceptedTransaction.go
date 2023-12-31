package transaction

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/crypto"
	"github.com/Dharitri-org/sme-dharitri/data"
	"github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/Dharitri-org/sme-dharitri/hashing"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/sharding"
	logger "github.com/Dharitri-org/sme-logger"
)

var _ process.TxValidatorHandler = (*InterceptedTransaction)(nil)
var _ process.InterceptedData = (*InterceptedTransaction)(nil)

// InterceptedTransaction holds and manages a transaction based struct with extended functionality
type InterceptedTransaction struct {
	tx                     *transaction.Transaction
	protoMarshalizer       marshal.Marshalizer
	signMarshalizer        marshal.Marshalizer
	hasher                 hashing.Hasher
	keyGen                 crypto.KeyGenerator
	singleSigner           crypto.SingleSigner
	pubkeyConv             core.PubkeyConverter
	coordinator            sharding.Coordinator
	hash                   []byte
	feeHandler             process.FeeHandler
	whiteListerVerifiedTxs process.WhiteListHandler
	argsParser             process.ArgumentsParser
	chainID                []byte
	minTransactionVersion  uint32
	rcvShard               uint32
	sndShard               uint32
	isForCurrentShard      bool
}

// NewInterceptedTransaction returns a new instance of InterceptedTransaction
func NewInterceptedTransaction(
	txBuff []byte,
	protoMarshalizer marshal.Marshalizer,
	signMarshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	keyGen crypto.KeyGenerator,
	signer crypto.SingleSigner,
	pubkeyConv core.PubkeyConverter,
	coordinator sharding.Coordinator,
	feeHandler process.FeeHandler,
	whiteListerVerifiedTxs process.WhiteListHandler,
	argsParser process.ArgumentsParser,
	chainID []byte,
	minTxVersion uint32,
) (*InterceptedTransaction, error) {

	if txBuff == nil {
		return nil, process.ErrNilBuffer
	}
	if check.IfNil(protoMarshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(signMarshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(keyGen) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(signer) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(pubkeyConv) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(coordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(feeHandler) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(whiteListerVerifiedTxs) {
		return nil, process.ErrNilWhiteListHandler
	}
	if check.IfNil(argsParser) {
		return nil, process.ErrNilArgumentParser
	}
	if len(chainID) == 0 {
		return nil, process.ErrInvalidChainID
	}
	if minTxVersion == 0 {
		return nil, process.ErrInvalidTransactionVersion
	}

	tx, err := createTx(protoMarshalizer, txBuff)
	if err != nil {
		return nil, err
	}

	inTx := &InterceptedTransaction{
		tx:                     tx,
		protoMarshalizer:       protoMarshalizer,
		signMarshalizer:        signMarshalizer,
		hasher:                 hasher,
		singleSigner:           signer,
		pubkeyConv:             pubkeyConv,
		keyGen:                 keyGen,
		coordinator:            coordinator,
		feeHandler:             feeHandler,
		whiteListerVerifiedTxs: whiteListerVerifiedTxs,
		argsParser:             argsParser,
		chainID:                chainID,
		minTransactionVersion:  minTxVersion,
	}

	err = inTx.processFields(txBuff)
	if err != nil {
		return nil, err
	}

	return inTx, nil
}

func createTx(marshalizer marshal.Marshalizer, txBuff []byte) (*transaction.Transaction, error) {
	tx := &transaction.Transaction{}
	err := marshalizer.Unmarshal(tx, txBuff)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// CheckValidity checks if the received transaction is valid (not nil fields, valid sig and so on)
func (inTx *InterceptedTransaction) CheckValidity() error {
	err := inTx.integrity(inTx.tx)
	if err != nil {
		return err
	}

	whiteListedVerified := inTx.whiteListerVerifiedTxs.IsWhiteListed(inTx)
	if !whiteListedVerified {
		err = inTx.verifySig(inTx.tx)
		if err != nil {
			return err
		}

		err = inTx.verifyIfRelayedTx(inTx.tx)
		if err != nil {
			return err
		}

		inTx.whiteListerVerifiedTxs.Add([][]byte{inTx.Hash()})
	}

	return nil
}

func (inTx *InterceptedTransaction) verifyIfRelayedTx(tx *transaction.Transaction) error {
	funcName, userTxArgs, err := inTx.argsParser.ParseCallData(string(tx.Data))
	if err != nil {
		return nil
	}
	if core.RelayedTransaction != funcName {
		return nil
	}

	if len(userTxArgs) != 1 {
		return process.ErrInvalidArguments
	}

	userTx, err := createTx(inTx.signMarshalizer, userTxArgs[0])
	if err != nil {
		return err
	}

	if !bytes.Equal(userTx.SndAddr, tx.RcvAddr) {
		return process.ErrRelayedTxBeneficiaryDoesNotMatchReceiver
	}

	err = inTx.integrity(userTx)
	if err != nil {
		return err
	}

	err = inTx.verifySig(userTx)
	if err != nil {
		return err
	}

	if len(userTx.Data) == 0 {
		return nil
	}

	funcName, _, err = inTx.argsParser.ParseCallData(string(userTx.Data))
	if err != nil {
		return nil
	}

	// recursive relayed transactions are not allowed
	if core.RelayedTransaction == funcName {
		return process.ErrRecursiveRelayedTxIsNotAllowed
	}

	return nil
}

func (inTx *InterceptedTransaction) processFields(txBuff []byte) error {
	inTx.hash = inTx.hasher.Compute(string(txBuff))

	inTx.sndShard = inTx.coordinator.ComputeId(inTx.tx.SndAddr)
	emptyAddr := make([]byte, len(inTx.tx.RcvAddr))
	inTx.rcvShard = inTx.coordinator.ComputeId(inTx.tx.RcvAddr)
	if bytes.Equal(inTx.tx.RcvAddr, emptyAddr) {
		inTx.rcvShard = inTx.sndShard
	}

	isForCurrentShardRecv := inTx.rcvShard == inTx.coordinator.SelfId()
	isForCurrentShardSender := inTx.sndShard == inTx.coordinator.SelfId()
	inTx.isForCurrentShard = isForCurrentShardRecv || isForCurrentShardSender

	return nil
}

// integrity checks for not nil fields and negative value
func (inTx *InterceptedTransaction) integrity(tx *transaction.Transaction) error {
	if tx.Version < inTx.minTransactionVersion {
		return process.ErrInvalidTransactionVersion
	}
	if !bytes.Equal(tx.ChainID, inTx.chainID) {
		return process.ErrInvalidChainID
	}
	if tx.Signature == nil {
		return process.ErrNilSignature
	}
	if tx.RcvAddr == nil {
		return process.ErrNilRcvAddr
	}
	if tx.SndAddr == nil {
		return process.ErrNilSndAddr
	}
	if tx.Value == nil {
		return process.ErrNilValue
	}
	if tx.Value.Sign() < 0 {
		return process.ErrNegativeValue
	}
	if len(inTx.tx.RcvUserName) > 0 && len(inTx.tx.RcvUserName) != inTx.hasher.Size() {
		return process.ErrInvalidUserNameLength
	}
	if len(inTx.tx.SndUserName) > 0 && len(inTx.tx.SndUserName) != inTx.hasher.Size() {
		return process.ErrInvalidUserNameLength
	}

	return inTx.feeHandler.CheckValidityTxValues(tx)
}

// verifySig checks if the tx is correctly signed
func (inTx *InterceptedTransaction) verifySig(tx *transaction.Transaction) error {
	buffCopiedTx, err := tx.GetDataForSigning(inTx.pubkeyConv, inTx.signMarshalizer)
	if err != nil {
		return err
	}

	senderPubKey, err := inTx.keyGen.PublicKeyFromByteArray(tx.SndAddr)
	if err != nil {
		return err
	}

	err = inTx.singleSigner.Verify(senderPubKey, buffCopiedTx, tx.Signature)
	if err != nil {
		return err
	}

	return nil
}

// ReceiverShardId returns the receiver shard id
func (inTx *InterceptedTransaction) ReceiverShardId() uint32 {
	return inTx.rcvShard
}

// IsForCurrentShard returns true if this transaction is meant to be processed by the node from this shard
func (inTx *InterceptedTransaction) IsForCurrentShard() bool {
	return inTx.isForCurrentShard
}

// Transaction returns the transaction pointer that actually holds the data
func (inTx *InterceptedTransaction) Transaction() data.TransactionHandler {
	return inTx.tx
}

// Hash gets the hash of this transaction
func (inTx *InterceptedTransaction) Hash() []byte {
	return inTx.hash
}

// SenderShardId returns the transaction sender shard id
func (inTx *InterceptedTransaction) SenderShardId() uint32 {
	return inTx.sndShard
}

// Nonce returns the transaction nonce
func (inTx *InterceptedTransaction) Nonce() uint64 {
	return inTx.tx.Nonce
}

// SenderAddress returns the transaction sender address
func (inTx *InterceptedTransaction) SenderAddress() []byte {
	return inTx.tx.SndAddr
}

// Fee returns the estimated cost of the transaction
func (inTx *InterceptedTransaction) Fee() *big.Int {
	return inTx.feeHandler.ComputeFee(inTx.tx)
}

// Type returns the type of this intercepted data
func (inTx *InterceptedTransaction) Type() string {
	return "intercepted tx"
}

// String returns the transaction's most important fields as string
func (inTx *InterceptedTransaction) String() string {
	return fmt.Sprintf("sender=%s, nonce=%d, value=%s, recv=%s",
		logger.DisplayByteSlice(inTx.tx.SndAddr),
		inTx.tx.Nonce,
		inTx.tx.Value.String(),
		logger.DisplayByteSlice(inTx.tx.RcvAddr),
	)
}

// Identifiers returns the identifiers used in requests
func (inTx *InterceptedTransaction) Identifiers() [][]byte {
	return [][]byte{inTx.hash}
}

// IsInterfaceNil returns true if there is no value under the interface
func (inTx *InterceptedTransaction) IsInterfaceNil() bool {
	return inTx == nil
}
