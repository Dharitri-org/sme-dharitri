package validator

import (
	"net/http"

	"github.com/Dharitri-org/sme-dharitri/api/errors"
	"github.com/Dharitri-org/sme-dharitri/api/shared"
	"github.com/Dharitri-org/sme-dharitri/api/wrapper"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/gin-gonic/gin"
)

const statisticsPath = "/statistics"

// FacadeHandler interface defines methods that can be used by the gin webserver
type FacadeHandler interface {
	ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error)
	IsInterfaceNil() bool
}

// Routes defines validators' related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, statisticsPath, Statistics)
}

func getFacade(c *gin.Context) (FacadeHandler, bool) {
	facadeObj, ok := c.Get("facade")
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrNilAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return nil, false
	}

	facade, ok := facadeObj.(FacadeHandler)
	if !ok {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: errors.ErrInvalidAppContext.Error(),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return nil, false
	}

	return facade, true
}

// Statistics will return the validation statistics for all validators
func Statistics(c *gin.Context) {
	facade, ok := getFacade(c)
	if !ok {
		return
	}

	valStats, err := facade.ValidatorStatisticsApi()
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  shared.ReturnCodeRequestError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"statistics": valStats},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}
