package alert

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"straas.io/base/logger"
	"straas.io/base/timeutil"
	"straas.io/sauron"
)

const (
	lastForNS = "plugin-lastfor"
)

var (
	log = logger.Get()
)

/*
lastfor returns a callback function to fill with key
lastfor allow two kinds of operations:
  lastfor(a > 10, "3m")("aaa.bbb.cc")
  lastfor(a, ">", 10, "3m")("aaa.bbb.cc")

*/

// LastForResult records result of lastfor
type LastForResult struct {
	Trigger bool
	Desc    string
}

// NewLastFor creates lastfor plugin
func NewLastFor(clock timeutil.Clock) sauron.Plugin {
	return &lastForPlugin{
		clock: clock,
	}
}

type lastForPlugin struct {
	clock timeutil.Clock
}

type lastForInfo struct {
	LastMatched int64 // unixtimestamp
}

func (p *lastForPlugin) Name() string {
	return "lastFor"
}

// Run run the lastfor,
func (p *lastForPlugin) Run(ctx sauron.PluginContext) error {
	if ctx.ArgLen() == 2 {
		return p.runBool(ctx)
	}
	if ctx.ArgLen() == 4 {
		return p.runCompare(ctx)
	}
	return fmt.Errorf("number of arguments should be 2 or 4")
}

// runBool runs with a bool value to indicate whether match or not
func (p *lastForPlugin) runBool(ctx sauron.PluginContext) error {
	matched, err := ctx.ArgBoolean(0)
	if err != nil {
		return err
	}
	durStr, err := ctx.ArgString(1)
	if err != nil {
		return err
	}
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		return err
	}
	term := "matched"
	if !matched {
		term = "unmatched"
	}
	return p.lastFor(matched, dur, term, ctx)
}

// runCompare runs with a compare expression (2 float and one operator)
// to indicate whether match or not
func (p *lastForPlugin) runCompare(ctx sauron.PluginContext) error {
	vLeft, err := ctx.ArgFloat(0)
	if err != nil {
		return err
	}
	op, err := ctx.ArgString(1)
	if err != nil {
		return err
	}
	vRight, err := ctx.ArgFloat(2)
	if err != nil {
		return err
	}
	durStr, err := ctx.ArgString(3)
	if err != nil {
		return err
	}
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		return err
	}

	match, err := opMatch(vLeft, op, vRight)
	if err != nil {
		return err
	}
	sLeft := strconv.FormatFloat(vLeft, 'f', 2, 64)
	sRight := strconv.FormatFloat(vRight, 'f', 2, 64)
	return p.lastFor(match, dur, fmt.Sprintf("%s %s %s", sLeft, op, sRight), ctx)
}

func (p *lastForPlugin) lastFor(matched bool, dur time.Duration, term string,
	ctx sauron.PluginContext) error {

	if dur < 0 {
		return fmt.Errorf("negative duration %v is not allowed", dur)
	}
	// trigger decides whether trigger alert or not
	trigger := false

	updater := func(v interface{}) (interface{}, error) {
		if v == nil {
			v = &lastForInfo{}
		}
		info := v.(*lastForInfo)
		trigger = p.updateStatus(matched, dur, info)
		return info, nil
	}
	callback := func(cbctx sauron.PluginContext) error {
		key, err := cbctx.ArgString(0)
		if err != nil {
			return err
		}
		info := &lastForInfo{}
		if err := cbctx.Store().Update(lastForNS, key, info, updater); err != nil {
			return err
		}
		// return a JSON
		cbctx.Return(&LastForResult{
			Trigger: trigger,
			Desc:    fmt.Sprintf("%s %s", key, term),
		})
		return nil
	}

	// return a callback function
	ctx.Return(sauron.FuncReturn(callback))
	return nil
}

// updateStatus updates status and decides whether trigger alert or not
func (p *lastForPlugin) updateStatus(match bool, threshold time.Duration, info *lastForInfo) bool {
	if !match {
		info.LastMatched = 0
		return false
	}
	// following are matched
	// threshold is zero, fire immediately
	if threshold == 0 {
		return true
	}
	now := p.clock.Now()
	// first meet the condition, record only
	if info.LastMatched == 0 {
		info.LastMatched = now.Unix()
		return false
	}
	lmTime := time.Unix(info.LastMatched, 0)
	dur := now.Sub(lmTime)
	if dur >= threshold {
		return true
	}
	log.Infof("last for %d seconds", dur/time.Second)
	return false
}

func (p *lastForPlugin) HelpMsg() string {
	return "<NO HELP MSG>"
}

func opMatch(vLeft float64, op string, vRight float64) (bool, error) {
	if math.IsInf(vLeft, 0) || math.IsInf(vRight, 0) {
		return false, nil
	}
	switch op {
	case ">":
		return vLeft > vRight, nil
	case ">=":
		return vLeft >= vRight, nil
	case "<":
		return vLeft < vRight, nil
	case "<=":
		return vLeft <= vRight, nil
	case "=":
		return vLeft == vRight, nil
	default:
		return false, fmt.Errorf("illegal operator %s", op)
	}
}
