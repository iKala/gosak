package alert

import (
	"math"
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
	s.eng = core.NewEngine(store, core.NewOutput(false))
	s.eng.AddPlugin(s.plugin)
}

func (s *lastforTestSuite) TestUpdateStatusNotMatched() {
	info := &lastForInfo{
		LastMatched: s.clock.Now().Unix(),
	}

	trigger := s.plugin.updateStatus(false, time.Minute, info)
	s.False(trigger)
	s.Equal(info.LastMatched, int64(0))
}

func (s *lastforTestSuite) TestUpdateStatusRightNow() {
	info := &lastForInfo{
		LastMatched: s.clock.Now().Unix(),
	}
	trigger := s.plugin.updateStatus(true, 0, info)
	s.True(trigger)
}

func (s *lastforTestSuite) TestUpdateStatusFirstMet() {
	info := &lastForInfo{
		LastMatched: 0,
	}
	trigger := s.plugin.updateStatus(true, time.Minute, info)
	s.False(trigger)
	s.Equal(info.LastMatched, s.clock.Now().Unix())
}

func (s *lastforTestSuite) TestUpdateStatusNotYet() {
	t1 := time.Now().Unix()
	info := &lastForInfo{
		LastMatched: t1,
	}
	s.clock.Incr(time.Minute)
	trigger := s.plugin.updateStatus(true, time.Hour, info)
	s.False(trigger)
	s.Equal(info.LastMatched, t1)
}

func (s *lastforTestSuite) TestUpdateStatusTrigger() {
	t1 := time.Now().Unix()
	info := &lastForInfo{
		LastMatched: t1,
	}
	s.clock.Incr(time.Hour)
	trigger := s.plugin.updateStatus(true, time.Minute, info)
	s.True(trigger)
	s.Equal(info.LastMatched, t1)
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

	match, err = opMatch(math.Inf(1), ">", 10)
	s.NoError(err)
	s.False(match)
	match, err = opMatch(math.Inf(1), "<", 10)
	s.NoError(err)
	s.False(match)
	match, err = opMatch(math.Inf(1), "=", 20)
	s.NoError(err)
	s.False(match)

	match, err = opMatch(1, ">", math.Inf(-1))
	s.NoError(err)
	s.False(match)
	match, err = opMatch(1, "<", math.Inf(-1))
	s.NoError(err)
	s.False(match)
	match, err = opMatch(1, "=", math.Inf(-1))
	s.NoError(err)
	s.False(match)

	match, err = opMatch(math.Inf(1), ">", math.Inf(-1))
	s.NoError(err)
	s.False(match)
	match, err = opMatch(math.Inf(1), "<", math.Inf(-1))
	s.NoError(err)
	s.False(match)
	match, err = opMatch(math.Inf(1), "=", math.Inf(-1))
	s.NoError(err)
	s.False(match)
}

func (s *lastforTestSuite) TestIllegalInput() {
	s.Error(s.run(`lastFor("xxx", "0s")`))
	s.Error(s.run(`lastFor(true, 3)`))
}

func (s *lastforTestSuite) TestBool() {
	s.NoError(s.run(`
		r = lastFor(true, "0s")("xxx")
		a = r.Desc
		b = r.Trigger
	`))
	v1, _ := s.eng.Get("a")
	v2, _ := s.eng.Get("b")

	s.Equal("xxx matched", v1)
	s.Equal(true, v2)
}

func (s *lastforTestSuite) TestBoolNotMatched() {
	s.NoError(s.run(`
		r = lastFor(7 < 5, "0s")("xxx")
		a = r.Desc
		b = r.Trigger
	`))
	v1, _ := s.eng.Get("a")
	v2, _ := s.eng.Get("b")

	s.Equal("xxx unmatched", v1)
	s.Equal(false, v2)
}

func (s *lastforTestSuite) TestComparsion() {
	s.NoError(s.run(`
		r = lastFor(4.2, ">=", 3.3, "0s")("xxx")
		a = r.Desc
		b = r.Trigger
	`))
	v1, _ := s.eng.Get("a")
	v2, _ := s.eng.Get("b")

	s.Equal("xxx 4.20 >= 3.30", v1)
	s.Equal(true, v2)
}

func (s *lastforTestSuite) TestComparsionNotMatched() {
	s.NoError(s.run(`
		r = lastFor(7, "<", 5, "0s")("xxx")
		a = r.Desc
		b = r.Trigger
	`))
	v1, _ := s.eng.Get("a")
	v2, _ := s.eng.Get("b")

	s.Equal("xxx 7.00 < 5.00", v1)
	s.Equal(false, v2)
}

func (s *lastforTestSuite) run(script string) error {
	meta := sauron.JobMeta{
		Script: script,
	}
	s.eng.SetJobMeta(meta)
	return s.eng.Run()
}
