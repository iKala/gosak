package core

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"straas.io/sauron/util"
)

const (
	testCacheNS    = "xxx"
	testCacheKey   = "yyy"
	testCacheValue = "aaa"
)

func TestClockSuite(t *testing.T) {
	suite.Run(t, new(cacheTestSuite))
}

type cacheTestSuite struct {
	suite.Suite
	cache *localCacheImpl
	clock util.FakeClock
}

func (s *cacheTestSuite) SetupTest() {
	s.clock = util.NewFakeClock()
	s.cache = NewCache(100, s.clock).(*localCacheImpl)
}

func (s *cacheTestSuite) TestWithTTL() {
	called := 0
	gen := func() (interface{}, error) {
		called++
		return testCacheValue, nil
	}
	for i := 0; i < 5; i++ {
		v, err := s.cache.Get(testCacheNS, testCacheKey, 0, gen)
		s.NoError(err)
		s.Equal(v, testCacheValue)
	}

	// forward years.
	s.clock.Incr(365 * 24 * time.Hour)

	// forward a year still reachable
	for i := 0; i < 5; i++ {
		v, err := s.cache.Get(testCacheNS, testCacheKey, 0, gen)
		s.NoError(err)
		s.Equal(v, testCacheValue)
	}

	s.Equal(called, 1)
}

func (s *cacheTestSuite) TestCacheMiss() {
	called := 0
	gen := func() (interface{}, error) {
		called++
		return testCacheValue, nil
	}

	// add value
	_, _ = s.cache.Get(testCacheNS, testCacheKey, 0, gen)
	s.Equal(called, 1)

	s.cache.cache.Remove(`["xxx","yyy"]`)

	_, _ = s.cache.Get(testCacheNS, testCacheKey, 0, gen)
	s.Equal(called, 2)
}

func (s *cacheTestSuite) TestExipre() {
	called := 0
	gen := func() (interface{}, error) {
		called++
		return testCacheValue, nil
	}

	// add value
	_, _ = s.cache.Get(testCacheNS, testCacheKey, time.Hour, gen)
	s.Equal(called, 1)

	// not exipre
	s.clock.Incr(59 * time.Minute)
	_, _ = s.cache.Get(testCacheNS, testCacheKey, 0, gen)
	s.Equal(called, 1)

	// exipre
	s.clock.Incr(2 * time.Minute)
	_, _ = s.cache.Get(testCacheNS, testCacheKey, 0, gen)
	s.Equal(called, 2)
}

func (s *cacheTestSuite) TestGenError() {
	gen := func() (interface{}, error) {
		return nil, errors.New("some error")
	}

	_, err := s.cache.Get(testCacheNS, testCacheKey, 0, gen)
	s.Error(err)
}
