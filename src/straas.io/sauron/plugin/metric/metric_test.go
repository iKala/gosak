package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"straas.io/base/timeutil"
	exmocks "straas.io/external/mocks"
	"straas.io/sauron"
	"straas.io/sauron/mocks"
)

const (
	testNow    = "2016-03-01T15:04:05+08:00"
	testEnv    = "some-prod"
	testModule = "test-module"
	testName   = "test-name"
	testField  = "test-field"
	testOp     = "sum"
	testSDur   = "48h"
	testEDur   = "1m"
)

func TestMetricSuite(t *testing.T) {
	suite.Run(t, new(metricTestSuite))
}

type metricTestSuite struct {
	suite.Suite
	plugin *metricQueryPlugin
	mockES *exmocks.Elasticsearch
	clock  timeutil.FakeClock
}

func (s *metricTestSuite) SetupTest() {
	s.clock = timeutil.NewFakeClock()
	s.mockES = &exmocks.Elasticsearch{}
	s.plugin = &metricQueryPlugin{
		es:    s.mockES,
		clock: s.clock,
	}
}

func (s *metricTestSuite) TestIndices() {
	start, _ := time.Parse(time.RFC3339, "2016-02-04T15:04:05+08:00")
	end, _ := time.Parse(time.RFC3339, "2016-02-05T15:04:05+08:00")
	end2, _ := time.Parse(time.RFC3339, "2016-05-05T15:04:05+08:00")

	indices, err := getIndices(start, end)

	s.NoError(err)
	s.Equal(indices, []string{
		"metric-2016.02",
	})

	indices, err = getIndices(start, end2)
	s.NoError(err)
	s.Equal(indices, []string{
		"metric-2016.02",
		"metric-2016.03",
		"metric-2016.04",
		"metric-2016.05",
	})

	// illegal range
	_, err = getIndices(end, start)
	s.Error(err)
}

func (s *metricTestSuite) TestArgument() {
	now, _ := time.Parse(time.RFC3339, testNow)
	s.clock.SetNow(now)

	ctx := &mocks.PluginContext{}
	ctx.On("JobMeta").Return(sauron.JobMeta{Env: testEnv}).Once()
	ctx.On("ArgString", 0).Return(testModule, nil).Once()
	ctx.On("ArgString", 1).Return(testName, nil).Once()
	ctx.On("ArgString", 2).Return(testField, nil).Once()
	ctx.On("ArgString", 3).Return(testOp, nil).Once()
	ctx.On("ArgString", 4).Return(testSDur, nil).Once()
	ctx.On("ArgString", 5).Return(testEDur, nil).Once()
	ctx.On("Return", 6.0).Return(nil).Once()

	result := 6.0
	s.mockES.On("Scalar",
		[]string{"metric-2016.02", "metric-2016.03"},
		`env=some-prod AND module=test-module AND name=test-name`,
		"@timestamp",
		now.Add(-48*time.Hour),
		now.Add(-1*time.Minute),
		testField,
		testOp,
		true,
	).Return(&result, nil).Once()

	err := s.plugin.Run(ctx)
	s.NoError(err)
	s.mockES.AssertExpectations(s.T())
	ctx.AssertExpectations(s.T())
}
