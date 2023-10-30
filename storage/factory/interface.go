package factory

import (
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/block/bootstrapStorage"
	"github.com/Dharitri-org/sme-dharitri/storage"
)

// BootstrapDataProviderHandler defines which actions should be done for loading bootstrap data from the boot storer
type BootstrapDataProviderHandler interface {
	LoadForPath(persisterFactory storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error)
	GetStorer(storer storage.Storer) (process.BootStorer, error)
	IsInterfaceNil() bool
}
