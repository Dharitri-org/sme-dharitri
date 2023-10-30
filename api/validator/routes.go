package validator

import (
	"net/http"

	"github.com/Dharitri-org/sme-dharitri/api/errors"
	"github.com/Dharitri-org/sme-dharitri/api/shared"
	"github.com/Dharitri-org/sme-dharitri/api/wrapper"
	"github.com/Dharitri-org/sme-dharitri/data/state"
	"github.com/gin-gonic/gin"
)

// ValidatorsStatisticsApiHandler interface defines methods that can be used from `dharitriFacade` context variable
type ValidatorsStatisticsApiHandler interface {
	ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error)
	IsInterfaceNil() bool
}

// Routes defines validators' related routes
func Routes(router *wrapper.RouterWrapper) {
	router.RegisterHandler(http.MethodGet, "/statistics", Statistics)
}

// Statistics will return the validation statistics for all validators
func Statistics(c *gin.Context) {
	ef, ok := c.MustGet("dharitriFacade").(ValidatorsStatisticsApiHandler)
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

	valStats, err := ef.ValidatorStatisticsApi()
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
