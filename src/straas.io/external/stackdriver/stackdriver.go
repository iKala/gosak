package stackdriver

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	monitoring "google.golang.org/api/monitoring/v3"

	"straas.io/external"
)

// New creates a stackdriver instance
func New(client *http.Client) (external.Stackdriver, error) {
	srv, err := monitoring.New(client)
	if err != nil {
		return nil, err
	}
	return &stackdriverImpl{
		tss: monitoring.NewProjectsTimeSeriesService(srv),
	}, nil
}

type stackdriverImpl struct {
	tss *monitoring.ProjectsTimeSeriesService
}

func (s *stackdriverImpl) List(project, filter, op string,
	start, end time.Time) ([]external.Point, error) {
	op = strings.ToUpper(op)

	reduce, align := "", ""
	// check op
	switch op {
	case "COUNT", "SUM", "MIN", "MAX":
		reduce = fmt.Sprintf("REDUCE_%s", op)
		align = fmt.Sprintf("ALIGN_%s", op)
	case "avg":
		reduce = "REDUCE_MEAN"
		align = "ALIGN_MEAN"
	default:
		return nil, fmt.Errorf("illegal op %s", op)
	}

	call := s.tss.List(fmt.Sprintf("projects/%s", project))
	call.
		AggregationAlignmentPeriod("60s").
		AggregationCrossSeriesReducer(reduce).
		AggregationPerSeriesAligner(align).
		IntervalStartTime(start.Format(time.RFC3339)).
		IntervalEndTime(end.Format(time.RFC3339)).
		Filter(filter)
	resp, err := call.Do()
	if err != nil {
		return nil, err
	}

	result := []external.Point{}
	for _, t := range resp.TimeSeries {
		// fmt.Println("value_type", t.ValueType)
		for _, p := range t.Points {
			// no enough data
			if p.Value == nil || p.Interval == nil {
				continue
			}
			v := p.Value.DoubleValue
			// conver int64 to float64
			if t.ValueType == "INT64" {
				v = float64(p.Value.Int64Value)
			}
			st, err := time.Parse(time.RFC3339, p.Interval.StartTime)
			if err != nil {
				return nil, err
			}
			ed, err := time.Parse(time.RFC3339, p.Interval.EndTime)
			if err != nil {
				return nil, err
			}

			result = append(result, external.Point{
				Start: st,
				End:   ed,
				Value: v,
			})
		}
	}
	return result, nil
}
