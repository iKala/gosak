package metric

import (
	"flag"
	"fmt"
	"strings"
	"time"

	elastic "gopkg.in/olivere/elastic.v3"

	"straas.io/sauron"
)

const (
	// max query time range
	maxQueryTimeRange = 90 * 24 * time.Hour // 3 month
	// query stat name
	aggrName = "stat"
)

var (
	indexPrefix = flag.String("metricIndexPrefix", "metric", "elasticsearch index prefix")
	datePattern = flag.String("metricDatePattern", "2006.01", "date pattern of eleasicsearch index")
	timeField   = flag.String("metricTimeField", "@timestamp", "metric time field")
	// for testing purpose
	timeNow = time.Now
)

// NewQuery create a metric query plugin
func NewQuery(client *elastic.Client) sauron.Plugin {
	return &metricQueryPlugin{
		client: client,
	}
}

// TODO: leverage LUR cache for better performance
type metricQueryPlugin struct {
	client *elastic.Client
}

func (p *metricQueryPlugin) Name() string {
	return "mquery"
}

func (p *metricQueryPlugin) Run(ctx sauron.PluginContext) error {
	modulePattern, err := ctx.ArgString(0)
	if err != nil {
		return err
	}
	namePattern, err := ctx.ArgString(1)
	if err != nil {
		return err
	}
	field, err := ctx.ArgString(2)
	if err != nil {
		return err
	}
	op, err := ctx.ArgString(3)
	if err != nil {
		return err
	}
	sduration, err := ctx.ArgString(4)
	if err != nil {
		return err
	}
	eduration, err := ctx.ArgString(5)
	if err != nil {
		return err
	}

	// query and set return value
	value, err := p.doQuery(
		timeNow(),
		ctx.JobMeta().Env,
		modulePattern,
		namePattern,
		field,
		op,
		sduration,
		eduration,
	)

	if err != nil {
		return err
	}
	if value == nil {
		// TBD: nil or zero ?
		return ctx.Return(nil)
	}
	return ctx.Return(*value)
}

func (p *metricQueryPlugin) HelpMsg() string {
	return "<NO HELP MSG>"
}

// doSerach for testing purpose
var doSearch = func(client *elastic.Client, source *elastic.SearchSource, indices []string) (*elastic.SearchResult, error) {
	search := client.Search()
	search.Index(indices...)

	return search.SearchSource(source).Do()
}

func (p *metricQueryPlugin) doQuery(now time.Time, env, modulePattern, namePattern,
	field, op, sduration, eduration string) (*float64, error) {

	source, indices, err := makeQuery(now, env, modulePattern, namePattern,
		field, op, sduration, eduration)
	if err != nil {
		return nil, err
	}

	// do query
	result, err := doSearch(p.client, source, indices)
	if err != nil {
		return nil, fmt.Errorf("query fail %v", err)
	}
	return handleResult(result, op)
}

func makeQuery(now time.Time, env, modulePattern, namePattern,
	field, op, sduration, eduration string) (*elastic.SearchSource, []string, error) {

	// make op case insensitive
	op = strings.ToLower(op)
	// check op
	switch op {
	case "count", "sum", "avg", "min", "max":
	default:
		return nil, nil, fmt.Errorf("illegal op %s", op)
	}

	start, err := time.ParseDuration(sduration)
	if err != nil {
		return nil, nil, err
	}
	end, err := time.ParseDuration(eduration)
	if err != nil {
		return nil, nil, err
	}
	st := now.Add(time.Duration(-start))
	en := now.Add(time.Duration(-end))
	if en.Sub(st) > maxQueryTimeRange {
		return nil, nil, fmt.Errorf("exceed max query range %v", maxQueryTimeRange)
	}
	// get elastic query indices
	indices, err := getIndices(st, en)
	if err != nil {
		return nil, nil, err
	}

	// make query request
	query := elastic.NewRangeQuery(*timeField).Gte(st).Lte(en)
	// perform wildcard search
	filter := elastic.NewQueryStringQuery(
		fmt.Sprintf(`env=%s AND module=%s AND name=%s`, env, modulePattern, namePattern),
	).AnalyzeWildcard(true)

	aggr := elastic.NewExtendedStatsAggregation().Field(field)
	source := elastic.NewSearchSource().Size(0)
	source = source.Query(elastic.NewBoolQuery().Must(query, filter))
	source = source.Aggregation(aggrName, aggr)

	return source, indices, nil
}

// getIndices generates search indices according to time range and flag
func getIndices(start, end time.Time) ([]string, error) {
	// TODO: handle too large range to protect elasticsearch
	// querying with lots of indices cost a large amount of memory
	result := []string{}
	cur := start.UTC().Truncate(24 * time.Hour)
	preIdx := ""
	for cur.Before(end) {
		idx := fmt.Sprintf("%s-%s", *indexPrefix, cur.Format(*datePattern))
		// bypass same index
		if idx != preIdx {
			preIdx = idx
			result = append(result, idx)
		}
		// to next day
		// golang does not have API for next month
		// so just add day by day for simple
		cur = cur.Add(24 * time.Hour)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no indices available betwen %s ~ %s",
			start.Format("2006.01.02"), end.Format("2006.01.02"))
	}
	return result, nil
}

// handleResult converts aggr result to float pointer or return error
func handleResult(result *elastic.SearchResult, op string) (*float64, error) {
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
