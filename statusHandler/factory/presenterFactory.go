package factory

import (
	"github.com/Dharitri-org/sme-dharitri/statusHandler/presenter"
	"github.com/Dharitri-org/sme-dharitri/statusHandler/view"
)

type presenterFactory struct {
}

// NewPresenterFactory is responsible for creating a new presenter factory object
func NewPresenterFactory() *presenterFactory {
	presenterFactoryObject := presenterFactory{}

	return &presenterFactoryObject
}

// Create returns an presenter object that will hold presenter in the system
func (pf *presenterFactory) Create() view.Presenter {
	presenterStatusHandler := presenter.NewPresenterStatusHandler()

	return presenterStatusHandler
}
