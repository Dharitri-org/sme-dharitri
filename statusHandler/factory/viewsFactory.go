package factory

import (
	"github.com/Dharitri-org/sme-dharitri/statusHandler"
	"github.com/Dharitri-org/sme-dharitri/statusHandler/view"
	"github.com/Dharitri-org/sme-dharitri/statusHandler/view/termuic"
)

type viewsFactory struct {
	presenter view.Presenter
}

// NewViewsFactory is responsible for creating a new viewers factory object
func NewViewsFactory(presenter view.Presenter) (*viewsFactory, error) {
	if presenter == nil || presenter.IsInterfaceNil() {
		return nil, statusHandler.ErrNilPresenterInterface
	}

	return &viewsFactory{
		presenter,
	}, nil
}

// Create returns an view slice that will hold all views in the system
func (wf *viewsFactory) Create() ([]Viewer, error) {
	views := make([]Viewer, 0)

	termuiConsole, err := wf.createTermuiConsole()
	if err != nil {
		return nil, err
	}
	views = append(views, termuiConsole)

	return views, nil
}

func (wf *viewsFactory) createTermuiConsole() (*termuic.TermuiConsole, error) {
	termuiConsole, err := termuic.NewTermuiConsole(wf.presenter)
	if err != nil {
		return nil, err
	}

	return termuiConsole, nil
}
