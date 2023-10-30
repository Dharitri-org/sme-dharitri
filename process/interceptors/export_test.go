package interceptors

import "github.com/Dharitri-org/sme-dharitri/process"

func (mdi *MultiDataInterceptor) Topic() string {
	return mdi.topic
}

func (mdi *MultiDataInterceptor) InterceptedDebugHandler() process.InterceptedDebugger {
	mdi.mutInterceptedDebugHandler.RLock()
	defer mdi.mutInterceptedDebugHandler.RUnlock()

	return mdi.interceptedDebugHandler
}

func (sdi *SingleDataInterceptor) Topic() string {
	return sdi.topic
}

func (sdi *SingleDataInterceptor) InterceptedDebugHandler() process.InterceptedDebugger {
	sdi.mutInterceptedDebugHandler.RLock()
	defer sdi.mutInterceptedDebugHandler.RUnlock()

	return sdi.interceptedDebugHandler
}
