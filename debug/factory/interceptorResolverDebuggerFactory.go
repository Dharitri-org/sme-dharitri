package factory

import (
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/debug/resolver"
)

// NewInterceptorResolverDebuggerFactory will instantiate an InterceptorResolverDebugHandler based on the provided config
func NewInterceptorResolverDebuggerFactory(config config.InterceptorResolverDebugConfig) (InterceptorResolverDebugHandler, error) {
	if !config.Enabled {
		return resolver.NewDisabledInterceptorResolver(), nil
	}

	return resolver.NewInterceptorResolver(config)
}
