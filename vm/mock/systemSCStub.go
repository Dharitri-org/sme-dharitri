package mock

import (
	vmcommon "github.com/Dharitri-org/sme-vm-common"
)

// SystemSCStub -
type SystemSCStub struct {
	ExecuteCalled func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode
	ValueOfCalled func(key interface{}) interface{}
}

// Execute -
func (s *SystemSCStub) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if s.ExecuteCalled != nil {
		return s.ExecuteCalled(args)
	}
	return 0
}

// ValueOf -
func (s *SystemSCStub) ValueOf(key interface{}) interface{} {
	if s.ValueOfCalled != nil {
		return s.ValueOfCalled(key)
	}
	return nil
}

// IsInterfaceNil -
func (s *SystemSCStub) IsInterfaceNil() bool {
	return s == nil
}
