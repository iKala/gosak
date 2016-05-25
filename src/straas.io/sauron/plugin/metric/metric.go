package metric

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"straas.io/base/timeutil"
	"straas.io/external"
	"straas.io/sauron"
)

const (
	// max query time range
	maxQueryTimeRange = 90 * 24 * time.Hour // 3 month
)

var (
	indexPrefix = flag.String("metricIndexPrefix", "metric", "elasticsearch index prefix")
	datePattern = flag.String("metricDatePattern", "2006.01", "date pattern of eleasicsearch index")
	timeField   = flag.String("metricTimeField", "@timestamp", "metric time field")
)

// NewQuery create a metric query plugin
func NewQuery(es external.Elasticsearch, clock timeutil.Clock) sauron.Plugin {
	return &metricQueryPlugin{
		es:    es,
		clock: clock,
	}
}

// TODO: leverage LUR cache for better performance
type metricQueryPlugin struct {
	es    external.Elasticsearch
	clock timeutil.Clock
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
		p.clock.Now(),
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

func (p *metricQueryPlugin) doQuery(now time.Time, env, modulePattern, namePattern,
	field, op, sduration, eduration string) (*float64, error) {

	start, err := time.ParseDuration(sduration)
	if err != nil {
		return nil, err
	}
	end, err := time.ParseDuration(eduration)
	if err != nil {
		return nil, err
	}
	st := now.Add(time.Duration(-start))
	en := now.Add(time.Duration(-end))
	if en.Sub(st) > maxQueryTimeRange {
		return nil, fmt.Errorf("exceed max query range %v", maxQueryTimeRange)
	}
	// get elastic query indices
	indices, err := getIndices(st, en)
	if err != nil {
		return nil, err
	}

	result, err := p.es.Scalar(
		indices,
		fmt.Sprintf(`env:%s AND module:%s AND name:%s`, env,
			escape(modulePattern), escape(namePattern)),
		*timeField,
		st, en,
		field,
		op,
		true,
	)
	if err != nil {
		return nil, err
	}
	if result == nil {
		v := float64(0.0)
		result = &v
	}
	return result, nil
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

// escape handles name with numeric char, numberic char must be quoted
func escape(s string) string {
	// alreay quoted
	if strings.HasPrefix(s, `"`) {
		return s
	}
	if !strings.ContainsAny(s, "1234567890") {
		return s
	}
	return fmt.Sprintf(`"%s"`, s)
}
