package alert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"straas.io/base/timeutil"
	"straas.io/sauron"
	"straas.io/sauron/core"
)

func TestLastForSuite(t *testing.T) {
	suite.Run(t, new(lastforTestSuite))
}

type lastforTestSuite struct {
	suite.Suite
	plugin *lastForPlugin
	clock  timeutil.FakeClock
	eng    sauron.Engine
}

func (s *lastforTestSuite) SetupTest() {
	s.clock = timeutil.NewFakeClock()
	s.plugin = &lastForPlugin{
		clock: s.clock,
	}
	store, _ := core.NewStore()
	s.eng = core.NewEngine(store)
	s.eng.AddPlugin(s.plugin)
}

func (s *lastforTestSuite) TestUpdateStatusNotMatched() {
	info := &lastForInfo{
		lastMatched: s.clock.Now().Unix(),
	}

	trigger := s.plugin.updateStatus(false, time.Minute, info)
	s.False(trigger)
	s.Equal(info.lastMatched, int64(0))
}

func (s *lastforTestSuite) TestUpdateStatusRightNow() {
	info := &lastForInfo{
		lastMatched: s.clock.Now().Unix(),
	}
	trigger := s.plugin.updateStatus(true, 0, info)
	s.True(trigger)
}

func (s *lastforTestSuite) TestUpdateStatusFirstMet() {
	info := &lastForInfo{
		lastMatched: 0,
	}
	trigger := s.plugin.updateStatus(true, time.Minute, info)
	s.False(trigger)
	s.Equal(info.lastMatched, s.clock.Now().Unix())
}

func (s *lastforTestSuite) TestUpdateStatusNotYet() {
	t1 := time.Now().Unix()
	info := &lastForInfo{
		lastMatched: t1,
	}
	s.clock.Incr(time.Minute)
	trigger := s.plugin.updateStatus(true, time.Hour, info)
	s.False(trigger)
	s.Equal(info.lastMatched, t1)
}

func (s *lastforTestSuite) TestUpdateStatusTrigger() {
	t1 := time.Now().Unix()
	info := &lastForInfo{
		lastMatched: t1,
	}
	s.clock.Incr(time.Hour)
	trigger := s.plugin.updateStatus(true, time.Minute, info)
	s.True(trigger)
	s.Equal(info.lastMatched, t1)
}

func (s *lastforTestSuite) TestOperator() {
	match, err := opMatch(10, "xx", 20)
	s.Error(err)

	match, err = opMatch(20, ">", 10)
	s.NoError(err)
	s.True(match)
	match, err = opMatch(10, ">", 10)
	s.NoError(err)
	s.False(match)
	match, err = opMatch(10, ">", 20)
	s.NoError(err)
	s.False(match)

	match, err = opMatch(20, ">=", 10)
	s.NoError(err)
	s.True(match)
	match, err = opMatch(10, ">=", 10)
	s.NoError(err)
	s.True(match)
	match, err = opMatch(10, ">=", 20)
	s.NoError(err)
	s.False(match)

	match, err = opMatch(20, "=", 10)
	s.NoError(err)
	s.False(match)
	match, err = opMatch(10, "=", 10)
	s.NoError(err)
	s.True(match)
	match, err = opMatch(10, "=", 20)
	s.NoError(err)
	s.False(match)

	match, err = opMatch(20, "<", 10)
	s.NoError(err)
	s.False(match)
	match, err = opMatch(10, "<", 10)
	s.NoError(err)
	s.False(match)
	match, err = opMatch(10, "<", 20)
	s.NoError(err)
	s.True(match)

	match, err = opMatch(20, "<=", 10)
	s.NoError(err)
	s.False(match)
	match, err = opMatch(10, "<=", 10)
	s.NoError(err)
	s.True(match)
	match, err = opMatch(10, "<=", 20)
	s.NoError(err)
	s.True(match)

}

func (s *lastforTestSuite) TestIllegalInput() {
	s.Error(s.run(`lastFor("xxx", "0s")`))
	s.Error(s.run(`lastFor(true, 3)`))
}

func (s *lastforTestSuite) TestBool() {
	s.NoError(s.run(`
		a = lastFor(true, "0s")("xxx")
	`))
	v, _ := s.eng.Get("a")

	s.Equal("xxx matched", v)
}

func (s *lastforTestSuite) TestBoolNotMatched() {
	s.NoError(s.run(`
		a = lastFor(7 < 5, "0s")("xxx")
	`))
	v, _ := s.eng.Get("a")

	s.Equal("", v)
}

func (s *lastforTestSuite) TestComparsion() {
	s.NoError(s.run(`
		a = lastFor(4.2, ">=", 3.3, "0s")("xxx")
	`))
	v, _ := s.eng.Get("a")

	s.Equal("xxx 4.200000 >= 3.300000", v)
}

func (s *lastforTestSuite) TestComparsionNotMatched() {
	s.NoError(s.run(`
		a = lastFor(7, "<", 5, "0s")("xxx")
	`))
	v, _ := s.eng.Get("a")

	s.Equal("", v)
}

func (s *lastforTestSuite) run(script string) error {
	meta := sauron.JobMeta{
		Script: script,
	}
	s.eng.SetJobMeta(meta)
	return s.eng.Run()
}
