package elasticsearch

import (
	"fmt"
	"strings"
	"time"

	elastic "gopkg.in/olivere/elastic.v3"

	"straas.io/external"
)

const (
	aggrName = "stat"
)

// New creates an elasticsearch instance
func New(client *elastic.Client) external.Elasticsearch {
	return &esImpl{
		client: client,
	}
}

type esImpl struct {
	client *elastic.Client
}

// Scalar queries es to scalar number
func (s *esImpl) Scalar(indices []string, queryStr, timeField string, start, end time.Time,
	field, op string, wildcard bool) (*float64, error) {

	// make op case insensitive
	op = strings.ToLower(op)
	// check op
	switch op {
	case "count", "sum", "avg", "min", "max":
	default:
		return nil, fmt.Errorf("illegal op %s", op)
	}

	// make query request
	query := elastic.NewRangeQuery(timeField).Gte(start).Lte(end)

	// perform wildcard search
	filter := elastic.NewQueryStringQuery(queryStr).AnalyzeWildcard(wildcard)
	aggr := elastic.NewExtendedStatsAggregation().Field(field)
	source := elastic.NewSearchSource()
	source = source.Query(elastic.NewBoolQuery().Must(query, filter))
	source = source.Aggregation(aggrName, aggr)

	search := s.client.Search()
	search.Index(indices...)

	result, err := search.SearchSource(source).Do()
	if err != nil {
		return nil, err
	}

	if result.TimedOut {
		return nil, fmt.Errorf("query timeout")
	}
	stats, ok := result.Aggregations.Stats(aggrName)
	if !ok {
		return nil, fmt.Errorf("no result aggregation (probably connection problem)")
	}
	switch op {
	case "min":
		return stats.Min, nil
	case "max":
		return stats.Max, nil
	case "avg":
		return stats.Avg, nil
	case "sum":
		return stats.Sum, nil
	case "count":
		fv := float64(stats.Count)
		return &fv, nil
	default:
		return nil, fmt.Errorf("no such field %s", op)
	}
}
