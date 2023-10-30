package factory

import (
	"github.com/Dharitri-org/sme-dharitri/config"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/core/statistics/softwareVersion"
)

type softwareVersionFactory struct {
	statusHandler core.AppStatusHandler
	config        config.SoftwareVersionConfig
}

// NewSoftwareVersionFactory is responsible for creating a new software version factory object
func NewSoftwareVersionFactory(
	statusHandler core.AppStatusHandler,
	config config.SoftwareVersionConfig,
) (*softwareVersionFactory, error) {
	if check.IfNil(statusHandler) {
		return nil, core.ErrNilAppStatusHandler
	}

	softwareVersionFactoryObject := &softwareVersionFactory{
		statusHandler: statusHandler,
		config:        config,
	}

	return softwareVersionFactoryObject, nil
}

// Create returns an software version checker object
func (svf *softwareVersionFactory) Create() (*softwareVersion.SoftwareVersionChecker, error) {
	stableTagProvider := softwareVersion.NewStableTagProvider(svf.config.StableTagLocation)
	softwareVersionChecker, err := softwareVersion.NewSoftwareVersionChecker(svf.statusHandler, stableTagProvider, svf.config.PollingIntervalInMinutes)

	return softwareVersionChecker, err
}
