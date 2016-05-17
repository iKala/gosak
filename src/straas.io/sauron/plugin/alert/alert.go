package alert

import (
	"flag"
	"fmt"
	"time"

	"straas.io/base/timeutil"
	"straas.io/sauron"
)

const (
	alertNS                      = "plugin-alert"
	minServerity sauron.Severity = 0
	maxServerity sauron.Severity = 2
	normal       sauron.Severity = 99
)

var (
	reNotifyInterval = flag.Duration("alertReNotify", 30*time.Minute, "alert re-notify interval")
)

// NewAlert creates lastfor plugin
func NewAlert(clock timeutil.Clock) sauron.Plugin {
	return &alertPlugin{
		clock: clock,
	}
}

type NotifyAction int

type alertPlugin struct {
	clock timeutil.Clock
}

type alertStatus struct {
	// p0, p1, p2
	Severity sauron.Severity
	// lastNotify is the last notify time in unix ts
	LastNotify int64

	FirstSeen int64
}

// notifier performs notification process
type notifier func(sauron.Severity, bool, string) error

// describer returns the description of the given severity level
type describer func(sauron.Severity) string

func (p *alertPlugin) Name() string {
	return "alert"
}

// Run run the lastfor,
func (p *alertPlugin) Run(ctx sauron.PluginContext) error {
	argLen := ctx.ArgLen()
	if argLen < 3 {
		return fmt.Errorf("not enough arguments")
	}
	if argLen > 5 {
		return fmt.Errorf("too many arguments")
	}
	name, err := ctx.ArgString(0)
	if err != nil {
		return err
	}
	if !ctx.IsCallable(1) {
		return fmt.Errorf("argument 1 is not callable(notify func)")
	}

	// alert key
	key := alertKey(ctx.JobMeta().JobID, name)
	// get most serious severity
	severity, descber, err := getSeverity(ctx, name)
	if err != nil {
		return err
	}
	notifier := p.getNotifier(ctx)
	updater := p.getUpdater(severity, descber, notifier)
	status := &alertStatus{}

	if err := ctx.Store().Update(alertNS, key, status, updater); err != nil {
		return err
	}
	return nil
}

func (p *alertPlugin) HelpMsg() string {
	return "<NO HELP MSG>"
}

func (p *alertPlugin) getNotifier(ctx sauron.PluginContext) notifier {
	// TODO: need better error handling
	return func(s sauron.Severity, resolve bool, desc string) error {
		if _, err := ctx.CallFunction(1, int(s), resolve, desc); err != nil {
			return err
		}
		return nil
	}
}

func (p *alertPlugin) getUpdater(severity sauron.Severity, descber describer,
	notify notifier) func(v interface{}) (interface{}, error) {

	return func(v interface{}) (interface{}, error) {
		if v == nil {
			v = &alertStatus{
				// default serverity is normal
				Severity: normal,
			}
		}
		status := v.(*alertStatus)
		preSeverity := status.Severity
		status.Severity = severity
		now := p.clock.Now()

		// Increase
		// e.g. normal to p1, or p2 to p0
		if severeThan(severity, preSeverity) {
			status.LastNotify = p.clock.Now().Unix()
			status.FirstSeen = status.LastNotify
			desc := fmt.Sprintf("Incident %s", descber(severity))
			if err := notify(severity, false, desc); err != nil {
				return nil, err
			}
			return status, nil
		}

		// Decrease
		// e.g. p1 to p2, p0 to normal
		if severeThan(preSeverity, severity) {
			desc := fmt.Sprintf("Resolve %s", descber(preSeverity))
			if err := notify(preSeverity, true, desc); err != nil {
				return nil, err
			}

			// not decrease to normal
			// e.g. p1 to p2, p0 to p1
			if severeThan(severity, normal) {
				status.LastNotify = p.clock.Now().Unix()
				status.FirstSeen = status.LastNotify
				desc = fmt.Sprintf("Incident %s", descber(severity))
				if err := notify(severity, false, desc); err != nil {
					return nil, err
				}
			}
			return status, nil
		}

		// keep severity level and not normal
		// e.g. p1 to p1
		if severity == preSeverity && severeThan(severity, normal) {
			// re-notify
			lastNotify := time.Unix(status.LastNotify, 0)
			if p.clock.Now().Sub(lastNotify) >= *reNotifyInterval {
				status.LastNotify = p.clock.Now().Unix()
				firstSeen := time.Unix(status.FirstSeen, 0)
				desc := fmt.Sprintf("Re-notify, incident %s has continued for %s",
					descber(severity),
					now.Sub(firstSeen))
				if err := notify(severity, false, desc); err != nil {
					return nil, err
				}
				return status, nil
			}
		}
		return nil, nil
	}
}

// severeThan returns whether the severity level of s1 is greater than s2
func severeThan(s1, s2 sauron.Severity) bool {
	return s1 < s2
}

// alertKey returns the key for alert
func alertKey(jobID, name string) string {
	return fmt.Sprintf("%s#%s", jobID, name)
}

// lastforKey returns the key for lastfor
func lastforKey(jobID, name string, severity sauron.Severity) string {
	return fmt.Sprintf("%s.P%d", alertKey(jobID, name), severity)
}

func getSeverity(ctx sauron.PluginContext, name string) (sauron.Severity, describer, error) {
	descs := []string{}
	targetSeverity := normal

	for sv := minServerity; sv <= maxServerity; sv++ {
		idx := int(sv) + 2 // first two arguments
		// index out of bound
		if idx >= ctx.ArgLen() {
			break
		}
		lkey := lastforKey(ctx.JobMeta().JobID, name, sv)
		v, err := ctx.CallFunction(idx, lkey)
		if err != nil {
			return 0, nil, err
		}
		result, ok := v.(*LastForResult)
		if !ok {
			return 0, nil, fmt.Errorf("return value is not an object, return:%v", v)
		}
		descs = append(descs, result.Desc)
		if result.Trigger && sv < targetSeverity {
			targetSeverity = sv
		}
	}
	// prepare describer
	descber := func(sv sauron.Severity) string {
		i := int(sv)
		if i >= 0 && i < len(descs) {
			return descs[i]
		}
		return ""
	}
	return targetSeverity, descber, nil
}
