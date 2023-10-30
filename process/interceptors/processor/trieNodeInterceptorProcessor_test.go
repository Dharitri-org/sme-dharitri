package processor_test

import (
	"testing"

	"github.com/Dharitri-org/sme-dharitri/core/check"
	"github.com/Dharitri-org/sme-dharitri/data/trie"
	"github.com/Dharitri-org/sme-dharitri/process"
	"github.com/Dharitri-org/sme-dharitri/process/interceptors/processor"
	"github.com/Dharitri-org/sme-dharitri/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieNodesInterceptorProcessor_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	tnip, err := processor.NewTrieNodesInterceptorProcessor(nil)
	assert.Nil(t, tnip)
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestNewTrieNodesInterceptorProcessor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	tnip, err := processor.NewTrieNodesInterceptorProcessor(testscommon.NewCacherMock())
	assert.Nil(t, err)
	assert.NotNil(t, tnip)
}

//------- Validate

func TestTrieNodesInterceptorProcessor_ValidateShouldWork(t *testing.T) {
	t.Parallel()

	tnip, _ := processor.NewTrieNodesInterceptorProcessor(testscommon.NewCacherMock())

	assert.Nil(t, tnip.Validate(nil, ""))
}

//------- Save

func TestTrieNodesInterceptorProcessor_SaveWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	tnip, _ := processor.NewTrieNodesInterceptorProcessor(testscommon.NewCacherMock())

	err := tnip.Save(nil, "", "")
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestTrieNodesInterceptorProcessor_SaveShouldPutInCacher(t *testing.T) {
	t.Parallel()

	putCalled := false
	cacher := &testscommon.CacherStub{
		PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
			putCalled = true
			return false
		},
	}
	tnip, _ := processor.NewTrieNodesInterceptorProcessor(cacher)

	err := tnip.Save(&trie.InterceptedTrieNode{}, "", "")
	assert.Nil(t, err)
	assert.True(t, putCalled)
}

//------- IsInterfaceNil

func TestTrieNodesInterceptorProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var tnip *processor.TrieNodeInterceptorProcessor
	assert.True(t, check.IfNil(tnip))
}
