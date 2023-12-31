package transaction_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apiErrors "github.com/Dharitri-org/sme-dharitri/api/errors"
	"github.com/Dharitri-org/sme-dharitri/api/middleware"
	"github.com/Dharitri-org/sme-dharitri/api/mock"
	"github.com/Dharitri-org/sme-dharitri/api/shared"
	"github.com/Dharitri-org/sme-dharitri/api/transaction"
	"github.com/Dharitri-org/sme-dharitri/api/wrapper"
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	tr "github.com/Dharitri-org/sme-dharitri/data/transaction"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type transactionResponseData struct {
	TxResp *transaction.TxResponse `json:"transaction,omitempty"`
}

type transactionResponse struct {
	Data  transactionResponseData `json:"data"`
	Error string                  `json:"error"`
	Code  string                  `json:"code"`
}

type sendMultipleTxsResponseData struct {
	TxsSent   int      `json:"txsSent"`
	TxsHashes []string `json:"txsHashes"`
}

type sendMultipleTxsResponse struct {
	Data  sendMultipleTxsResponseData `json:"data"`
	Error string                      `json:"error"`
	Code  string                      `json:"code"`
}

type sendSingleTxResponseData struct {
	TxHash string `json:"txHash"`
}

type sendSingleTxResponse struct {
	Data  sendSingleTxResponseData `json:"data"`
	Error string                   `json:"error"`
	Code  string                   `json:"code"`
}

type transactionCostResponseData struct {
	Cost uint64 `json:"txGasUnits"`
}

type transactionCostResponse struct {
	Data  transactionCostResponseData `json:"data"`
	Error string                      `json:"error"`
	Code  string                      `json:"code"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetTransaction_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("GET", "/transaction/hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestGetTransaction_WithCorrectHashShouldReturnTransaction(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := "10"
	txData := []byte("data")
	hash := "hash"
	facade := mock.Facade{
		GetTransactionHandler: func(hash string) (i *tr.ApiTransactionResult, e error) {
			return &tr.ApiTransactionResult{
				Sender:   sender,
				Receiver: receiver,
				Data:     txData,
				Value:    value,
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "/transaction/"+hash, nil)
	ws := startNodeServer(&facade)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := transactionResponse{}
	loadResponse(resp.Body, &response)

	txResp := response.Data.TxResp
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, sender, txResp.Sender)
	assert.Equal(t, receiver, txResp.Receiver)
	assert.Equal(t, value, txResp.Value)
	assert.Equal(t, txData, txResp.Data)
}

func TestGetTransaction_WithUnknownHashShouldReturnNil(t *testing.T) {
	sender := "sender"
	receiver := "receiver"
	value := "10"
	txData := []byte("data")
	wrongHash := "wronghash"
	facade := mock.Facade{
		GetTransactionHandler: func(hash string) (*tr.ApiTransactionResult, error) {
			if hash == wrongHash {
				return nil, errors.New("local error")
			}
			return &tr.ApiTransactionResult{
				Sender:   sender,
				Receiver: receiver,
				Data:     txData,
				Value:    value,
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "/transaction/"+wrongHash, nil)
	ws := startNodeServer(&facade)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := transactionResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Empty(t, txResp.Data)
}

func TestGetTransaction_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/transaction/empty", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := transactionResponse{}
	loadResponse(resp.Body, &txResp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, txResp.Error, apiErrors.ErrInvalidAppContext.Error())
}

func TestGetTransaction_ErrorWithExceededNumGoRoutines(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		GetThrottlerForEndpointCalled: func(_ string) (core.Throttler, bool) {
			return &mock.ThrottlerStub{
				CanProcessCalled: func() bool { return false },
			}, true
		},
	}
	ws := startNodeServer(&facade)

	req, _ := http.NewRequest("GET", "/transaction/eeee", nil)

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := transactionResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code)
	assert.True(t, strings.Contains(txResp.Error, apiErrors.ErrTooManyRequests.Error()))
	assert.Equal(t, string(shared.ReturnCodeSystemBusy), txResp.Code)
	assert.Empty(t, txResp.Data)
}

func TestSendTransaction_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("POST", "/transaction/send", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestSendTransaction_ErrorWithWrongFacade(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("POST", "/transaction/send", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendSingleTxResponse{}
	loadResponse(resp.Body, &txResp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, apiErrors.ErrInvalidAppContext.Error(), txResp.Error)
}

func TestSendTransaction_ErrorWithExceededNumGoRoutines(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		GetThrottlerForEndpointCalled: func(_ string) (core.Throttler, bool) {
			return &mock.ThrottlerStub{
				CanProcessCalled: func() bool { return false },
			}, true
		},
	}
	ws := startNodeServer(&facade)

	tx := transaction.SendTxRequest{}

	jsonBytes, _ := json.Marshal(tx)
	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendSingleTxResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code)
	assert.True(t, strings.Contains(txResp.Error, apiErrors.ErrTooManyRequests.Error()))
	assert.Equal(t, string(shared.ReturnCodeSystemBusy), txResp.Code)
	assert.Empty(t, txResp.Data)
}

func TestSendTransaction_WrongParametersShouldErrorOnValidation(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := "ishouldbeint"
	data := "data"

	facade := mock.Facade{}
	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(`{"sender":"%s", "receiver":"%s", "value":%s, "data":"%s"}`,
		sender,
		receiver,
		value,
		data,
	)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendSingleTxResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, txResp.Error, apiErrors.ErrValidation.Error())
	assert.Empty(t, txResp.Data)
}

func TestSendTransaction_ErrorWhenFacadeSendTransactionError(t *testing.T) {
	t.Parallel()
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	signature := "aabbccdd"
	errorString := "send transaction error"

	facade := mock.Facade{
		CreateTransactionHandler: func(_ uint64, _ string, _ string, _ string, _ uint64, _ uint64, _ []byte, _ string, _ string, _ uint32,
		) (*tr.Transaction, []byte, error) {
			return nil, nil, nil
		},
		SendBulkTransactionsHandler: func(txs []*tr.Transaction) (u uint64, err error) {
			return 0, errors.New(errorString)
		},
		ValidateTransactionHandler: func(tx *tr.Transaction) error {
			return nil
		},
	}
	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(`{"sender":"%s", "receiver":"%s", "value":"%s", "signature":"%s", "data":"%s"}`,
		sender,
		receiver,
		value,
		signature,
		data,
	)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendSingleTxResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, txResp.Error, errorString)
	assert.Empty(t, txResp.Data)
}

func TestSendTransaction_ReturnsSuccessfully(t *testing.T) {
	t.Parallel()
	nonce := uint64(1)
	sender := "sender"
	receiver := "receiver"
	value := big.NewInt(10)
	data := "data"
	signature := "aabbccdd"
	hexTxHash := "deadbeef"

	facade := mock.Facade{
		CreateTransactionHandler: func(_ uint64, _ string, _ string, _ string, _ uint64, _ uint64, _ []byte, _ string, _ string, _ uint32,
		) (*tr.Transaction, []byte, error) {
			txHash, _ := hex.DecodeString(hexTxHash)
			return nil, txHash, nil
		},
		SendBulkTransactionsHandler: func(txs []*tr.Transaction) (u uint64, err error) {
			return 1, nil
		},
		ValidateTransactionHandler: func(tx *tr.Transaction) error {
			return nil
		},
	}
	ws := startNodeServer(&facade)

	jsonStr := fmt.Sprintf(
		`{"nonce": %d, "sender": "%s", "receiver": "%s", "value": "%s", "signature": "%s", "data": "%s"}`,
		nonce,
		sender,
		receiver,
		value,
		signature,
		data,
	)

	req, _ := http.NewRequest("POST", "/transaction/send", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := sendSingleTxResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Empty(t, response.Error)
	assert.Equal(t, hexTxHash, response.Data.TxHash)
}

func TestSendMultipleTransactions_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("POST", "/transaction/send-multiple", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestSendMultipleTransactions_ErrorWithWrongFacade(t *testing.T) {
	t.Parallel()

	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("POST", "/transaction/send-multiple", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendMultipleTxsResponse{}
	loadResponse(resp.Body, &txResp)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Equal(t, apiErrors.ErrInvalidAppContext.Error(), txResp.Error)
}

func TestSendMultipleTransactions_ErrorWithExceededNumGoRoutines(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{
		GetThrottlerForEndpointCalled: func(_ string) (core.Throttler, bool) {
			return &mock.ThrottlerStub{
				CanProcessCalled: func() bool { return false },
			}, true
		},
	}
	ws := startNodeServer(&facade)

	tx0 := transaction.SendTxRequest{}
	txs := []*transaction.SendTxRequest{&tx0}

	jsonBytes, _ := json.Marshal(txs)
	req, _ := http.NewRequest("POST", "/transaction/send-multiple", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendMultipleTxsResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusTooManyRequests, resp.Code)
	assert.True(t, strings.Contains(txResp.Error, apiErrors.ErrTooManyRequests.Error()))
	assert.Equal(t, string(shared.ReturnCodeSystemBusy), txResp.Code)
	assert.Empty(t, txResp.Data)
}

func TestSendMultipleTransactions_WrongPayloadShouldErrorOnValidation(t *testing.T) {
	t.Parallel()

	facade := mock.Facade{}
	ws := startNodeServer(&facade)

	jsonStr := `{"wrong": json}`

	req, _ := http.NewRequest("POST", "/transaction/send-multiple", bytes.NewBuffer([]byte(jsonStr)))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txResp := sendMultipleTxsResponse{}
	loadResponse(resp.Body, &txResp)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, txResp.Error, apiErrors.ErrValidation.Error())
	assert.Empty(t, txResp.Data)
}

func TestSendMultipleTransactions_OkPayloadShouldWork(t *testing.T) {
	t.Parallel()

	createTxWasCalled := false
	sendBulkTxsWasCalled := false

	facade := mock.Facade{
		CreateTransactionHandler: func(_ uint64, _ string, _ string, _ string, _ uint64, _ uint64, _ []byte, _ string, _ string, _ uint32,
		) (*tr.Transaction, []byte, error) {
			createTxWasCalled = true
			return &tr.Transaction{}, make([]byte, 0), nil
		},
		SendBulkTransactionsHandler: func(txs []*tr.Transaction) (u uint64, e error) {
			sendBulkTxsWasCalled = true
			return 0, nil
		},
		ValidateTransactionHandler: func(tx *tr.Transaction) error {
			return nil
		},
	}
	ws := startNodeServer(&facade)

	tx0 := transaction.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      make([]byte, 0),
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}
	tx1 := tx0
	tx1.Sender = "sender2"
	txs := []*transaction.SendTxRequest{&tx0, &tx1}

	jsonBytes, _ := json.Marshal(txs)

	req, _ := http.NewRequest("POST", "/transaction/send-multiple", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txCostResp := sendMultipleTxsResponse{}
	loadResponse(resp.Body, &txCostResp)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.True(t, createTxWasCalled)
	assert.True(t, sendBulkTxsWasCalled)
}

func TestComputeTransactionGasLimit_NilContextShouldError(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)

	req, _ := http.NewRequest("POST", "/transaction/cost", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
	response := shared.GenericAPIResponse{}
	loadResponse(resp.Body, &response)

	assert.Equal(t, shared.ReturnCodeInternalError, response.Code)
	assert.True(t, strings.Contains(response.Error, apiErrors.ErrNilAppContext.Error()))
}

func TestComputeTransactionGasLimit(t *testing.T) {
	t.Parallel()

	expectedGasLimit := uint64(37)

	facade := mock.Facade{
		CreateTransactionHandler: func(_ uint64, _ string, _ string, _ string, _ uint64, _ uint64, _ []byte, _ string, _ string, _ uint32,
		) (*tr.Transaction, []byte, error) {
			return &tr.Transaction{}, nil, nil
		},
		ComputeTransactionGasLimitHandler: func(tx *tr.Transaction) (uint64, error) {
			return expectedGasLimit, nil
		},
	}
	ws := startNodeServer(&facade)

	tx0 := transaction.SendTxRequest{
		Sender:    "sender1",
		Receiver:  "receiver1",
		Value:     "100",
		Data:      make([]byte, 0),
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: "",
	}

	jsonBytes, _ := json.Marshal(tx0)

	req, _ := http.NewRequest("POST", "/transaction/cost", bytes.NewBuffer(jsonBytes))

	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	txCostResp := transactionCostResponse{}
	loadResponse(resp.Body, &txCostResp)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, expectedGasLimit, txCostResp.Data.Cost)
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	if err != nil {
		logError(err)
	}
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func startNodeServer(handler transaction.FacadeHandler) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ginTransactionRoute := ws.Group("/transaction")
	if handler != nil {
		ginTransactionRoute.Use(middleware.WithFacade(handler))
	}
	transactionRoute, _ := wrapper.NewRouterWrapper("transaction", ginTransactionRoute, getRoutesConfig())
	transaction.Routes(transactionRoute)
	return ws
}

func startNodeServerWrongFacade() *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	ws.Use(func(c *gin.Context) {
		c.Set("facade", mock.WrongFacade{})
	})
	ginTransactionRoute := ws.Group("/transaction")
	transactionRoute, _ := wrapper.NewRouterWrapper("transaction", ginTransactionRoute, getRoutesConfig())
	transaction.Routes(transactionRoute)
	return ws
}

func getRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"transaction": {
				[]config.RouteConfig{
					{Name: "/send", Open: true},
					{Name: "/send-multiple", Open: true},
					{Name: "/cost", Open: true},
					{Name: "/:txhash", Open: true},
					{Name: "/:txhash/status", Open: true},
				},
			},
		},
	}
}
