package external_test

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/node/external"
	"github.com/Dharitri-org/sme-dharitri/node/mock"
	"github.com/Dharitri-org/sme-dharitri/process"
	vmcommon "github.com/Dharitri-org/sme-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestNewNodeApiResolver_NilSCQueryServiceShouldErr(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(nil, &mock.StatusMetricsStub{}, &mock.TransactionCostEstimatorMock{})

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilSCQueryService, err)
}

func TestNewNodeApiResolver_NilStatusMetricsShouldErr(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(&mock.SCQueryServiceStub{}, nil, &mock.TransactionCostEstimatorMock{})

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilStatusMetrics, err)
}

func TestNewNodeApiResolver_NilTransactionCostEstsimator(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(&mock.SCQueryServiceStub{}, &mock.StatusMetricsStub{}, nil)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilTransactionCostHandler, err)
}

func TestNewNodeApiResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(&mock.SCQueryServiceStub{}, &mock.StatusMetricsStub{}, &mock.TransactionCostEstimatorMock{})

	assert.Nil(t, err)
	assert.False(t, check.IfNil(nar))
}

func TestNodeApiResolver_GetDataValueShouldCall(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(&mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (vmOutput *vmcommon.VMOutput, e error) {
			wasCalled = true
			return &vmcommon.VMOutput{}, nil
		},
	},
		&mock.StatusMetricsStub{}, &mock.TransactionCostEstimatorMock{})

	_, _ = nar.ExecuteSCQuery(&process.SCQuery{
		ScAddress: []byte{0},
		FuncName:  "",
	})

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusMetricsMapWithoutP2PShouldBeCalled(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(
		&mock.SCQueryServiceStub{},
		&mock.StatusMetricsStub{
			StatusMetricsMapWithoutP2PCalled: func() map[string]interface{} {
				wasCalled = true
				return nil
			},
		},
		&mock.TransactionCostEstimatorMock{},
	)
	_ = nar.StatusMetrics().StatusMetricsMapWithoutP2P()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusP2pMetricsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(
		&mock.SCQueryServiceStub{},
		&mock.StatusMetricsStub{
			StatusP2pMetricsMapCalled: func() map[string]interface{} {
				wasCalled = true
				return nil
			},
		},
		&mock.TransactionCostEstimatorMock{},
	)
	_ = nar.StatusMetrics().StatusP2pMetricsMap()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusMetricsMapWhitoutP2PShouldBeCalled(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(
		&mock.SCQueryServiceStub{},
		&mock.StatusMetricsStub{
			StatusMetricsMapWithoutP2PCalled: func() map[string]interface{} {
				wasCalled = true
				return nil
			},
		},
		&mock.TransactionCostEstimatorMock{},
	)
	_ = nar.StatusMetrics().StatusMetricsMapWithoutP2P()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusP2PMetricsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(
		&mock.SCQueryServiceStub{},
		&mock.StatusMetricsStub{
			StatusP2pMetricsMapCalled: func() map[string]interface{} {
				wasCalled = true
				return nil
			},
		},
		&mock.TransactionCostEstimatorMock{},
	)
	_ = nar.StatusMetrics().StatusP2pMetricsMap()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_NetworkMetricsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(
		&mock.SCQueryServiceStub{},
		&mock.StatusMetricsStub{
			NetworkMetricsCalled: func() map[string]interface{} {
				wasCalled = true
				return nil
			},
		},
		&mock.TransactionCostEstimatorMock{},
	)
	_ = nar.StatusMetrics().NetworkMetrics()

	assert.True(t, wasCalled)
}
