package api

import (
	"net/http"
	"reflect"

	"github.com/Dharitri-org/sme-dharitri/api/address"
	"github.com/Dharitri-org/sme-dharitri/api/hardfork"
	"github.com/Dharitri-org/sme-dharitri/api/logs"
	"github.com/Dharitri-org/sme-dharitri/api/middleware"
	"github.com/Dharitri-org/sme-dharitri/api/network"
	"github.com/Dharitri-org/sme-dharitri/api/node"
	"github.com/Dharitri-org/sme-dharitri/api/transaction"
	valStats "github.com/Dharitri-org/sme-dharitri/api/validator"
	"github.com/Dharitri-org/sme-dharitri/api/vmValues"
	"github.com/Dharitri-org/sme-dharitri/api/wrapper"
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/marshal"
	logger "github.com/Dharitri-org/sme-logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gorilla/websocket"
	"gopkg.in/go-playground/validator.v8"
)

var log = logger.GetOrCreate("api")

type validatorInput struct {
	Name      string
	Validator validator.Func
}

// MiddlewareProcessor defines a processor used internally by the web server when processing requests
type MiddlewareProcessor interface {
	MiddlewareHandlerFunc() gin.HandlerFunc
	IsInterfaceNil() bool
}

// MainApiHandler interface defines methods that can be used from `dharitriFacade` context variable
type MainApiHandler interface {
	RestApiInterface() string
	RestAPIServerDebugMode() bool
	PprofEnabled() bool
	IsInterfaceNil() bool
}

type ginWriter struct {
}

func (gv *ginWriter) Write(p []byte) (n int, err error) {
	log.Debug("gin server", "message", string(p))

	return len(p), nil
}

type ginErrorWriter struct {
}

func (gev *ginErrorWriter) Write(p []byte) (n int, err error) {
	log.Debug("gin server", "error", string(p))

	return len(p), nil
}

// Start will boot up the api and appropriate routes, handlers and validators
func Start(dharitriFacade MainApiHandler, routesConfig config.ApiRoutesConfig, middleware MiddlewareProcessor) error {
	var ws *gin.Engine
	if !dharitriFacade.RestAPIServerDebugMode() {
		gin.DefaultWriter = &ginWriter{}
		gin.DefaultErrorWriter = &ginErrorWriter{}
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
	}
	ws = gin.Default()
	ws.Use(cors.Default())
	if !check.IfNil(middleware) {
		ws.Use(middleware.MiddlewareHandlerFunc())
	}

	err := registerValidators()
	if err != nil {
		return err
	}

	registerRoutes(ws, routesConfig, dharitriFacade)

	return ws.Run(dharitriFacade.RestApiInterface())
}

func registerRoutes(ws *gin.Engine, routesConfig config.ApiRoutesConfig, dharitriFacade middleware.DharitriHandler) {
	nodeRoutes := ws.Group("/node")
	wrappedNodeRouter, err := wrapper.NewRouterWrapper("node", nodeRoutes, routesConfig)
	if err == nil {
		node.Routes(wrappedNodeRouter)
	}

	addressRoutes := ws.Group("/address")
	wrappedAddressRouter, err := wrapper.NewRouterWrapper("address", addressRoutes, routesConfig)
	if err == nil {
		address.Routes(wrappedAddressRouter)
	}

	networkRoutes := ws.Group("/network")
	wrappedNetworkRoutes, err := wrapper.NewRouterWrapper("network", networkRoutes, routesConfig)
	if err == nil {
		network.Routes(wrappedNetworkRoutes)
	}

	txRoutes := ws.Group("/transaction")
	wrappedTransactionRouter, err := wrapper.NewRouterWrapper("transaction", txRoutes, routesConfig)
	if err == nil {
		transaction.Routes(wrappedTransactionRouter)
	}

	vmValuesRoutes := ws.Group("/vm-values")
	wrappedVmValuesRouter, err := wrapper.NewRouterWrapper("vm-values", vmValuesRoutes, routesConfig)
	if err == nil {
		vmValues.Routes(wrappedVmValuesRouter)
	}

	validatorRoutes := ws.Group("/validator")
	wrappedValidatorsRouter, err := wrapper.NewRouterWrapper("validator", validatorRoutes, routesConfig)
	if err == nil {
		valStats.Routes(wrappedValidatorsRouter)
	}

	hardforkRoutes := ws.Group("/hardfork")
	wrappedHardforkRouter, err := wrapper.NewRouterWrapper("hardfork", hardforkRoutes, routesConfig)
	if err == nil {
		hardfork.Routes(wrappedHardforkRouter)
	}

	apiHandler, ok := dharitriFacade.(MainApiHandler)
	if ok && apiHandler.PprofEnabled() {
		pprof.Register(ws)
	}

	if isLogRouteEnabled(routesConfig) {
		marshalizerForLogs := &marshal.GogoProtoMarshalizer{}
		registerLoggerWsRoute(ws, marshalizerForLogs)
	}
}

func isLogRouteEnabled(routesConfig config.ApiRoutesConfig) bool {
	logConfig, ok := routesConfig.APIPackages["log"]
	if !ok {
		return false
	}

	for _, cfg := range logConfig.Routes {
		if cfg.Name == "/log" && cfg.Open {
			return true
		}
	}

	return false
}

func registerValidators() error {
	validators := []validatorInput{
		{Name: "skValidator", Validator: skValidator},
	}
	for _, validatorFunc := range validators {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			err := v.RegisterValidation(validatorFunc.Name, validatorFunc.Validator)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func registerLoggerWsRoute(ws *gin.Engine, marshalizer marshal.Marshalizer) {
	upgrader := websocket.Upgrader{}

	ws.GET("/log", func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error(err.Error())
			return
		}

		ls, err := logs.NewLogSender(marshalizer, conn, log)
		if err != nil {
			log.Error(err.Error())
			return
		}

		ls.StartSendingBlocking()
	})
}

// skValidator validates a secret key from user input for correctness
func skValidator(
	_ *validator.Validate,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Type,
	_ reflect.Kind,
	_ string,
) bool {
	return true
}
