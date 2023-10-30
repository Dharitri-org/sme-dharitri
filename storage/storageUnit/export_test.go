package storageUnit

import (
	"github.com/Dharitri-org/sme-dharitri/storage"
)

func (u *Unit) GetBlomFilter() storage.BloomFilter {
	return u.bloomFilter
}
