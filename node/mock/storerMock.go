package mock

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
)

// StorerMock -
type StorerMock struct {
	mut  sync.Mutex
	data map[string][]byte
}

// RangeKeys -
func (sm *StorerMock) RangeKeys(_ func(_ []byte, _ []byte) bool) {
	panic("implement me")
}

// NewStorerMock -
func NewStorerMock() *StorerMock {
	return &StorerMock{
		data: make(map[string][]byte),
	}
}

// Put -
func (sm *StorerMock) Put(key, data []byte) error {
	sm.mut.Lock()
	defer sm.mut.Unlock()
	sm.data[string(key)] = data

	return nil
}

// Get -
func (sm *StorerMock) Get(key []byte) ([]byte, error) {
	sm.mut.Lock()
	defer sm.mut.Unlock()

	val, ok := sm.data[string(key)]
	if !ok {
		return nil, fmt.Errorf("key: %s not found", base64.StdEncoding.EncodeToString(key))
	}

	return val, nil
}

// GetFromEpoch -
func (sm *StorerMock) GetFromEpoch(key []byte, _ uint32) ([]byte, error) {
	return sm.Get(key)
}

// GetBulkFromEpoch -
func (sm *StorerMock) GetBulkFromEpoch(keys [][]byte, _ uint32) (map[string][]byte, error) {
	retValue := map[string][]byte{}
	for _, key := range keys {
		value, err := sm.Get(key)
		if err != nil {
			continue
		}
		retValue[string(key)] = value
	}

	return retValue, nil
}

// HasInEpoch -
func (sm *StorerMock) HasInEpoch(_ []byte, _ uint32) error {
	return errors.New("not implemented")
}

// SearchFirst -
func (sm *StorerMock) SearchFirst(_ []byte) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Close -
func (sm *StorerMock) Close() error {
	return nil
}

// Has -
func (sm *StorerMock) Has(_ []byte) error {
	return errors.New("not implemented")
}

// Remove -
func (sm *StorerMock) Remove(_ []byte) error {
	return errors.New("not implemented")
}

// ClearCache -
func (sm *StorerMock) ClearCache() {
}

// DestroyUnit -
func (sm *StorerMock) DestroyUnit() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *StorerMock) IsInterfaceNil() bool {
	return sm == nil
}
