package network

import (
	"net/http"

	"github.com/Dharitri-org/sme-dharitri/api/errors"
	"github.com/Dharitri-org/sme-dharitri/api/shared"
	"github.com/Dharitri-org/sme-dharitri/api/wrapper"
	"github.com/Dharitri-org/sme-dharitri/node/external"
	"github.com/gin-gonic/gin"
)

// FacadeHandler interface defines methods that can be used from `dharitriFacade` context variable
type FacadeHandler interface {
	StatusMetrics() external.StatusMetricsHandler
	IsInterfaceNil() bool
}

// Routes defines address related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, "/config", GetNetworkConfig)
	router.RegisterHandler(http.MethodGet, "/status", GetNetworkStatus)
}

// GetNetworkConfig returns metrics related to the network configuration (shard independent)
func GetNetworkConfig(c *gin.Context) {
	ef, ok := c.MustGet("dharitriFacade").(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	configMetrics := ef.StatusMetrics().ConfigMetrics()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"config": configMetrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

// GetNetworkStatus returns metrics related to the network status (shard specific)
func GetNetworkStatus(c *gin.Context) {
	ef, ok := c.MustGet("dharitriFacade").(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	networkMetrics := ef.StatusMetrics().NetworkMetrics()
	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"status": networkMetrics},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}
