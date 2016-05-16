package alert

import (
	"fmt"
	"time"

	"straas.io/base/timeutil"
	"straas.io/sauron"
)

const (
	lastForNS = "plugin-lastfor"
)

/*
lastfor returns a callback function to fill with key
lastfor allow two kinds of operations:
  lastfor(a > 10, "3m")("aaa.bbb.cc")
  lastfor(a, ">", 10, "3m")("aaa.bbb.cc")

alert(
  "cpu,                             // name
  notify("frontend", "backend"),    // action
  lastFor(expression, "3m"),        // P0
  lastFor(v1, "op", v2, "3m")       // P1
)
*/

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
	lastMatched int64 // unixtimestamp
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
	return fmt.Errorf("number of arguments is 2 or 4")
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
	return p.lastFor(match, dur, fmt.Sprintf("%f %s %f", vLeft, op, vRight), ctx)
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
func (p *lastForPlugin) updateStatus(match bool, dur time.Duration, info *lastForInfo) bool {
	if !match {
		info.lastMatched = 0
		return false
	}
	// following are matched
	// dur is zero, fire immediately
	if dur == 0 {
		return true
	}
	now := p.clock.Now()
	// first meet the condition, record only
	if info.lastMatched == 0 {
		info.lastMatched = now.Unix()
		return false
	}
	lmTime := time.Unix(info.lastMatched, 0)
	// TBD: allowed error
	if now.Sub(lmTime) >= dur {
		return true
	}
	return false
}

func (p *lastForPlugin) HelpMsg() string {
	return "<NO HELP MSG>"
}

func opMatch(vLeft float64, op string, vRight float64) (bool, error) {
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
