package stackdriver

import (
	"fmt"
	"strings"
	"time"

	"github.com/montanaflynn/stats"

	"straas.io/base/timeutil"
	"straas.io/external"
	"straas.io/sauron"
)

const (
	// max query time range
	maxQueryTimeRange = 90 * 24 * time.Hour // 3 month
)

// NewQuery create a metric query plugin
func NewQuery(proj2sd map[string]external.Stackdriver,
	env2proj map[string]string, clock timeutil.Clock) sauron.Plugin {
	return &squeryPlugin{
		proj2sd:  proj2sd,
		env2proj: env2proj,
		clock:    clock,
	}
}

// TODO: leverage LUR cache for better performance
type squeryPlugin struct {
	proj2sd  map[string]external.Stackdriver
	env2proj map[string]string
	clock    timeutil.Clock
}

func (p *squeryPlugin) Name() string {
	return "squery"
}

func (p *squeryPlugin) Run(ctx sauron.PluginContext) error {
	filter, err := ctx.ArgString(0)
	if err != nil {
		return err
	}
	op, err := ctx.ArgString(1)
	if err != nil {
		return err
	}
	sduration, err := ctx.ArgString(2)
	if err != nil {
		return err
	}
	eduration, err := ctx.ArgString(3)
	if err != nil {
		return err
	}

	project, ok := p.env2proj[ctx.JobMeta().Env]
	if !ok {
		return fmt.Errorf("fail to map env %s to project", ctx.JobMeta().Env)
	}

	// query and set return value
	value, err := p.doQuery(
		p.clock.Now(),
		project,
		filter,
		op,
		sduration,
		eduration,
	)

	if err != nil {
		return err
	}
	if value == nil {
		// return 0
		return ctx.Return(0)
	}
	return ctx.Return(*value)
}

func (p *squeryPlugin) HelpMsg() string {
	return "<NO HELP MSG>"
}

func (p *squeryPlugin) doQuery(now time.Time, project, filter,
	op, sduration, eduration string) (*float64, error) {

	op = strings.ToLower(op)
	switch op {
	case "sum", "avg", "min", "max":
	default:
		return nil, fmt.Errorf("illegal op %s", op)
	}

	sd, ok := p.proj2sd[project]
	if !ok {
		return nil, fmt.Errorf("unknown project %s", project)
	}
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
	// query stackdriver
	points, err := sd.List(project, filter, op, st, en)
	if err != nil {
		return nil, err
	}
	// no data
	if len(points) == 0 {
		return nil, nil
	}
	result := stat(op, points)
	return &result, nil
}

func stat(op string, points []external.Point) float64 {
	values := make([]float64, 0, len(points))
	for _, p := range points {
		values = append(values, p.Value)
	}

	// only empty array causes stats to report error
	// so just ignore it
	var result float64
	switch op {
	case "sum":
		result, _ = stats.Sum(values)
	case "avg":
		result, _ = stats.Mean(values)
	case "min":
		result, _ = stats.Min(values)
	case "max":
		result, _ = stats.Max(values)
	}
	return result
}
