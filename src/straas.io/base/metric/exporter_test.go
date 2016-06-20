package metric

import (
	"testing"
	"time"

	"github.com/csigo/metric"
	"github.com/stretchr/testify/suite"

	"straas.io/external/mocks"
)

func TestExporter(t *testing.T) {
	suite.Run(t, new(exporterTestSuite))
}

type exporterTestSuite struct {
	suite.Suite
}

// snapshot implements metric.Snapshot interface
type snapshot struct {
	pkg  string
	name string
	metric.CounterSnapshot
	metric.HistSnapshot
}

func (s *snapshot) Pkg() string {
	return s.pkg
}

func (s *snapshot) Name() string {
	return s.name
}

func (s *snapshot) HasHistogram() bool {
	return s.HistSnapshot != nil
}

func (s *exporterTestSuite) TestExportOnce() {
	lastUpdate := time.Now().Add(-4 * time.Minute)

	getSnapshot = func(pkg, name string) []metric.Snapshot {
		now := time.Now()

		c1 := &metric.MockCtrSnapshot{}
		c1.On("SliceIn", getInterval).Return([]metric.Bucket{
			metric.Bucket{
				Count: 2,
				Sum:   10,
				Avg:   5,
				Start: now.Add(-5 * time.Minute),
				End:   now.Add(-4 * time.Minute),
			},
			metric.Bucket{
				Count: 3,
				Sum:   30,
				Avg:   10,
				Start: now.Add(-4 * time.Minute),
				End:   now.Add(-3 * time.Minute),
			},
		})
		s1 := &snapshot{
			pkg:             "test",
			name:            "aaa.bbb1",
			CounterSnapshot: c1,
		}

		c2 := &metric.MockCtrSnapshot{}
		c2.On("HasHistogram").Return(false)
		c2.On("SliceIn", getInterval).Return([]metric.Bucket{
			metric.Bucket{
				Count: 2,
				Sum:   10,
				Avg:   5,
				Start: now.Add(-10 * time.Minute),
				End:   now.Add(-9 * time.Minute),
			},
			metric.Bucket{
				Count: 3,
				Sum:   30,
				Avg:   10,
				Start: now.Add(-9 * time.Minute),
				End:   now.Add(-8 * time.Minute),
			},
		})
		s2 := &snapshot{
			pkg:             "test",
			name:            "aaa.bbb2",
			CounterSnapshot: c2,
		}

		c3 := &metric.MockCtrSnapshot{}
		h3 := &metric.MockHistSnapshot{}
		c3.On("HasHistogram").Return(true)
		c3.On("SliceIn", getInterval).Return([]metric.Bucket{
			metric.Bucket{
				Count: 2,
				Sum:   40,
				Avg:   5,
				Start: now.Add(-4 * time.Minute),
				End:   now.Add(-3 * time.Minute),
			},
			metric.Bucket{
				Count: 3,
				Sum:   50,
				Avg:   10,
				Start: now.Add(-4 * time.Minute),
				End:   now.Add(-3 * time.Minute),
			},
		})
		h3.On("Percentiles", percentiles).Return([]float64{10, 20, 30}, int64(6))
		s3 := &snapshot{
			pkg:             "test",
			name:            "aaa.bbb3",
			CounterSnapshot: c3,
			HistSnapshot:    h3,
		}

		c4 := &metric.MockCtrSnapshot{}
		h4 := &metric.MockHistSnapshot{}
		c4.On("HasHistogram").Return(true)
		c4.On("SliceIn", getInterval).Return([]metric.Bucket{
			metric.Bucket{
				Count: 2,
				Sum:   40,
				Avg:   5,
				Start: now.Add(-10 * time.Minute),
				End:   now.Add(-9 * time.Minute),
			},
			metric.Bucket{
				Count: 3,
				Sum:   50,
				Avg:   10,
				Start: now.Add(-9 * time.Minute),
				End:   now.Add(-8 * time.Minute),
			},
		})
		s4 := &snapshot{
			pkg:             "test",
			name:            "aaa.bbb4",
			CounterSnapshot: c4,
			HistSnapshot:    h4,
		}

		return []metric.Snapshot{s1, s2, s3, s4}
	}

	testTag := "ttt"
	hostName = "test-host"
	mockFluent := mocks.Fluent{}
	mockFluent.On("Post", testTag, map[string]interface{}{
		"module": "test",
		"host":   "test-host",
		"name":   "aaa.bbb1",
		"count":  float64(3),
		"value":  float64(30),
		"avg":    float64(10),
	}).Once()
	mockFluent.On("Post", testTag, map[string]interface{}{
		"module": "test",
		"host":   "test-host",
		"name":   "aaa.bbb3",
		"count":  float64(3),
		"value":  float64(50),
		"avg":    float64(10),
	}).Once()
	mockFluent.On("Post", testTag, map[string]interface{}{
		"module": "test",
		"host":   "test-host",
		"name":   "aaa.bbb3.p50",
		"count":  float64(6),
		"value":  float64(60),
		"avg":    float64(10),
	}).Once()
	mockFluent.On("Post", testTag, map[string]interface{}{
		"module": "test",
		"host":   "test-host",
		"name":   "aaa.bbb3.p95",
		"count":  float64(6),
		"value":  float64(120),
		"avg":    float64(20),
	}).Once()
	mockFluent.On("Post", testTag, map[string]interface{}{
		"module": "test",
		"host":   "test-host",
		"name":   "aaa.bbb3.p99",
		"count":  float64(6),
		"value":  float64(180),
		"avg":    float64(30),
	}).Once()

	exp := createExporter(&mockFluent, testTag)
	exportOnce(lastUpdate, exp)
	mockFluent.AssertExpectations(s.T())
}
