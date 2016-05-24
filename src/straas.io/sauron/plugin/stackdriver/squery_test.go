package stackdriver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"straas.io/base/timeutil"
	"straas.io/external"
	exmocks "straas.io/external/mocks"
	"straas.io/sauron"
	"straas.io/sauron/mocks"
)

const (
	testNow     = "2016-03-01T15:04:05+08:00"
	testEnv     = "some-prod"
	testProject = "test-project"
	testFilter  = "aaa=10 AND bbb=5"
	testOp      = "sum"
	testSDur    = "5m"
	testEDur    = "1m"
)

func TestSquerySuite(t *testing.T) {
	suite.Run(t, new(squeryTestSuite))
}

type squeryTestSuite struct {
	suite.Suite
	plugin *squeryPlugin
	mockSD *exmocks.Stackdriver
	clock  timeutil.FakeClock
}

func (s *squeryTestSuite) SetupTest() {
	s.clock = timeutil.NewFakeClock()
	s.mockSD = &exmocks.Stackdriver{}
	s.plugin = &squeryPlugin{
		proj2sd: map[string]external.Stackdriver{
			testProject: s.mockSD,
		},
		clock: s.clock,
		env2proj: map[string]string{
			testEnv: testProject,
		},
	}
}

func (s *squeryTestSuite) TestArgument() {
	now, _ := time.Parse(time.RFC3339, testNow)
	s.clock.SetNow(now)

	ctx := &mocks.PluginContext{}
	ctx.On("JobMeta").Return(sauron.JobMeta{Env: testEnv}).Once()
	ctx.On("ArgString", 0).Return(testFilter, nil).Once()
	ctx.On("ArgString", 1).Return(testOp, nil).Once()
	ctx.On("ArgString", 2).Return(testSDur, nil).Once()
	ctx.On("ArgString", 3).Return(testEDur, nil).Once()
	ctx.On("Return", 8.8).Return(nil).Once()

	result := []external.Point{
		external.Point{
			Start: s.clock.Now(),
			End:   s.clock.Now(),
			Value: 3.3,
		},
		external.Point{
			Start: s.clock.Now(),
			End:   s.clock.Now(),
			Value: 5.5,
		},
	}
	s.mockSD.On("List",
		testProject,
		testFilter,
		testOp,
		now.Add(-5*time.Minute),
		now.Add(-1*time.Minute),
	).Return(result, nil).Once()

	err := s.plugin.Run(ctx)
	s.NoError(err)
	s.mockSD.AssertExpectations(s.T())
	ctx.AssertExpectations(s.T())
}

func (s *squeryTestSuite) TestStats() {
	points := []external.Point{
		external.Point{
			Start: s.clock.Now(),
			End:   s.clock.Now(),
			Value: 3.3,
		},
		external.Point{
			Start: s.clock.Now(),
			End:   s.clock.Now(),
			Value: 5.5,
		},
	}
	s.Equal(stat("sum", points), 8.8)
	s.Equal(stat("avg", points), 4.4)
	s.Equal(stat("max", points), 5.5)
	s.Equal(stat("min", points), 3.3)
}
