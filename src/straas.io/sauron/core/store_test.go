package core

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"straas.io/sauron"
)

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(storeTestSuite))
}

type testData struct {
	Name string
}

type storeTestSuite struct {
	suite.Suite
	store sauron.Store
}

func (s *storeTestSuite) SetupTest() {
	s.store, _ = NewStore()
}

func (s *storeTestSuite) TestGetSet() {
	v1 := &testData{Name: "v1"}

	err := s.store.Set("ns1", "key", v1)
	s.NoError(err)

	v := &testData{}
	ok, err := s.store.Get("ns1", "key", v)
	s.True(ok)
	s.NoError(err)
	s.Equal(v, v1)
}

func (s *storeTestSuite) TestNotFound() {
	v := &testData{}
	ok, err := s.store.Get("ns1", "key", v)
	s.False(ok)
	s.NoError(err)
}

func (s *storeTestSuite) TestDiffKey() {
	v1 := &testData{Name: "v1"}
	v2 := &testData{Name: "v2"}

	s.store.Set("ns1", "key1", v1)
	s.store.Set("ns1", "key2", v2)

	v := &testData{}
	ok, err := s.store.Get("ns1", "key1", v)
	s.True(ok)
	s.NoError(err)
	s.Equal(v, v1)

	v = &testData{}
	ok, err = s.store.Get("ns1", "key2", v)
	s.True(ok)
	s.NoError(err)
	s.Equal(v, v2)
}

func (s *storeTestSuite) TestDiffNamespace() {
	v1 := &testData{Name: "v1"}
	v2 := &testData{Name: "v2"}

	s.store.Set("ns1", "key", v1)
	s.store.Set("ns2", "key", v2)

	v := &testData{}
	ok, err := s.store.Get("ns1", "key", v)
	s.True(ok)
	s.NoError(err)
	s.Equal(v, v1)

	v = &testData{}
	ok, err = s.store.Get("ns2", "key", v)
	s.True(ok)
	s.NoError(err)
	s.Equal(v, v2)
}

func (s *storeTestSuite) TestUnmarshalFail() {
	v1 := &testData{Name: "v1"}

	err := s.store.Set("ns1", "key", v1)
	s.NoError(err)

	v := []string{}
	_, err = s.store.Get("ns1", "key", v)
	s.Error(err)
}
