package metric

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	elastic "gopkg.in/olivere/elastic.v3"

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
	testQuery  = `{
  "aggregations": {
    "stat": {
      "extended_stats": {
        "field": "test-field"
      }
    }
  },
  "query": {
    "bool": {
      "must": [
        {
          "range": {
            "@timestamp": {
              "from": "2016-02-28T15:04:05+08:00",
              "include_lower": true,
              "include_upper": true,
              "to": "2016-03-01T15:03:05+08:00"
            }
          }
        },
        {
          "query_string": {
            "analyze_wildcard": true,
            "query": "env=some-prod AND module=test-module AND name=test-name"
          }
        }
      ]
    }
  },
  "size": 0
}`
	testResult = `{
		"aggregations": {
		  	"stat": {
		    "count": 2,
		    "min": 3.0,
		    "max": 4.0,
		    "avg": 5.0,
		    "sum": 6.0
		  	}
	  	}
	}`
)

func TestMetricSuite(t *testing.T) {
	suite.Run(t, new(metricTestSuite))
}

type metricTestSuite struct {
	suite.Suite
	plugin *metricQueryPlugin
}

func (s *metricTestSuite) SetupTest() {
	s.plugin = &metricQueryPlugin{}
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

func (s *metricTestSuite) TestMakeQuerySuccess() {
	now, _ := time.Parse(time.RFC3339, testNow)

	source, indices, err := makeQuery(
		now,
		testEnv,
		testModule,
		testName,
		testField,
		testOp,
		testSDur,
		testEDur)
	s.NoError(err)
	s.Equal(indices, []string{"metric-2016.02", "metric-2016.03"})
	s.assertSource(source, testQuery)
}

func (s *metricTestSuite) TestCaseInsensitive() {
	_, _, err := makeQuery(
		time.Now(),
		testEnv,
		testModule,
		testName,
		testField,
		"SuM",
		testSDur,
		testEDur)
	s.NoError(err)
}

func (s *metricTestSuite) TestMakeQueryFail() {
	// start end exchange
	_, _, err := makeQuery(
		time.Now(),
		testEnv,
		testModule,
		testName,
		testField,
		testOp,
		testEDur,
		testSDur,
	)
	s.Error(err)

	// illegal OP
	_, _, err = makeQuery(
		time.Now(),
		testEnv,
		testModule,
		testName,
		testField,
		"xxx",
		testEDur,
		testSDur,
	)
	s.Error(err)

	// too large range
	_, _, err = makeQuery(
		time.Now(),
		testEnv,
		testModule,
		testName,
		testField,
		testOp,
		"3000h",
		testEDur,
	)
	s.Error(err)

	// illegal start format
	_, _, err = makeQuery(
		time.Now(),
		testEnv,
		testModule,
		testName,
		testField,
		testOp,
		"uuu",
		testEDur,
	)
	s.Error(err)

	// illegal end format
	_, _, err = makeQuery(
		time.Now(),
		testEnv,
		testModule,
		testName,
		testField,
		testOp,
		testSDur,
		"uuu",
	)
	s.Error(err)
}

func (s *metricTestSuite) TestHandleResult() {
	result := &elastic.SearchResult{}
	result.TimedOut = true

	_, err := handleResult(result, testOp)
	s.Error(err)

	result = &elastic.SearchResult{}
	json.Unmarshal([]byte(testResult), result)

	v, err := handleResult(result, "count")
	s.NoError(err)
	s.NotNil(v)
	s.Equal(float64(2), *v)

	v, err = handleResult(result, "min")
	s.NoError(err)
	s.NotNil(v)
	s.Equal(float64(3), *v)

	v, err = handleResult(result, "max")
	s.NoError(err)
	s.NotNil(v)
	s.Equal(float64(4), *v)

	v, err = handleResult(result, "avg")
	s.NoError(err)
	s.NotNil(v)
	s.Equal(float64(5), *v)

	v, err = handleResult(result, "sum")
	s.NoError(err)
	s.NotNil(v)
	s.Equal(float64(6), *v)

	// no such field
	_, err = handleResult(result, "XXX")
	s.Error(err)
}

func (s *metricTestSuite) TestArgument() {
	now, _ := time.Parse(time.RFC3339, testNow)
	timeNow = func() time.Time {
		return now
	}

	ctx := &mocks.PluginContext{}
	ctx.On("JobMeta").Return(sauron.JobMeta{Env: testEnv}).Once()
	ctx.On("ArgString", 0).Return(testModule, nil).Once()
	ctx.On("ArgString", 1).Return(testName, nil).Once()
	ctx.On("ArgString", 2).Return(testField, nil).Once()
	ctx.On("ArgString", 3).Return(testOp, nil).Once()
	ctx.On("ArgString", 4).Return(testSDur, nil).Once()
	ctx.On("ArgString", 5).Return(testEDur, nil).Once()
	ctx.On("Return", 6.0).Return(nil).Once()

	called := false
	doSearch = func(client *elastic.Client, source *elastic.SearchSource, indices []string) (*elastic.SearchResult, error) {
		called = true
		s.assertSource(source, testQuery)

		result := &elastic.SearchResult{}
		json.Unmarshal([]byte(testResult), result)
		return result, nil
	}

	err := s.plugin.Run(ctx)
	s.NoError(err)
	s.True(called)
	ctx.AssertExpectations(s.T())
}

func (s *metricTestSuite) assertSource(source *elastic.SearchSource, rawJSON string) {
	ss, err := source.Source()
	s.NoError(err)

	rawQuery, _ := json.Marshal(ss)
	query := map[string]interface{}{}
	expQuery := map[string]interface{}{}

	json.Unmarshal(rawQuery, &query)
	json.Unmarshal([]byte(testQuery), &expQuery)
	s.Equal(query, expQuery)
}
