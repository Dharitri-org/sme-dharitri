package data

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/stretchr/testify/assert"
)

func TestInitialSmartContract_OwnerBytes(t *testing.T) {
	t.Parallel()

	ownerBytes := []byte("owner bytes")
	isc := &InitialSmartContract{}
	isc.SetOwnerBytes(ownerBytes)
	recovered := isc.OwnerBytes()

	assert.Equal(t, ownerBytes, recovered)
}

func TestInitialSmartContract_VmTypeBytes(t *testing.T) {
	t.Parallel()

	vmTypeBytes := []byte("vm type bytes")
	isc := &InitialSmartContract{}
	isc.SetVmTypeBytes(vmTypeBytes)
	recovered := isc.VmTypeBytes()

	assert.Equal(t, vmTypeBytes, recovered)
}

func TestInitialSmartContract_Getters(t *testing.T) {
	t.Parallel()

	owner := "owner"
	filename := "filename"
	vmType := "vm type"
	initParams := "init parameters"
	scType := "type"
	version := "version"
	isc := &InitialSmartContract{
		Owner:          owner,
		Filename:       filename,
		VmType:         vmType,
		Type:           scType,
		InitParameters: initParams,
		Version:        version,
	}

	assert.False(t, check.IfNil(isc))
	assert.Equal(t, owner, isc.GetOwner())
	assert.Equal(t, filename, isc.GetFilename())
	assert.Equal(t, vmType, isc.GetVmType())
	assert.Equal(t, scType, isc.GetType())
	assert.Equal(t, initParams, isc.GetInitParameters())
	assert.Equal(t, version, isc.GetVersion())
}

func TestInitialSmartContract_AddressBytes(t *testing.T) {
	t.Parallel()

	ia := &InitialSmartContract{}
	addrBytes := []byte("address bytes")
	ia.AddAddressBytes(addrBytes)
	recovmoaAddrBytes := ia.AddressesBytes()[0]

	assert.Equal(t, addrBytes, recovmoaAddrBytes)
}

func TestInitialSmartContract_Address(t *testing.T) {
	t.Parallel()

	ia := &InitialSmartContract{}
	address := "address"
	ia.AddAddress(address)
	recovmoaAddress := ia.Addresses()[0]

	assert.Equal(t, address, recovmoaAddress)
}
