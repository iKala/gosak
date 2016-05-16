package alert

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"straas.io/base/timeutil"
	"straas.io/sauron"
	"straas.io/sauron/core"
	"straas.io/sauron/plugin/util"
)

const (
	testJobID = "test-job-id"
)

func TestAlertSuite(t *testing.T) {
	suite.Run(t, new(alertTestSuite))
}

type alertTestSuite struct {
	suite.Suite
	plugin       *alertPlugin
	notifyPlugin *util.TestPlugin
	notifies     [][]interface{}
	clock        timeutil.FakeClock
	eng          sauron.Engine
	store        sauron.Store
}

func (s *alertTestSuite) SetupTest() {
	s.clock = timeutil.NewFakeClock()
	s.plugin = &alertPlugin{
		clock: s.clock,
	}
	s.notifyPlugin = util.NewTestPlugin()
	s.notifyPlugin.PluginName = "notify"
	s.notifyPlugin.RunFunc = func(ctx sauron.PluginContext) error {
		sv, err := ctx.ArgInt(0)
		s.NoError(err)
		resolve, err := ctx.ArgBoolean(1)
		s.NoError(err)
		desc, err := ctx.ArgString(2)
		s.NoError(err)
		s.notifies = append(s.notifies, []interface{}{
			Severity(sv), resolve, desc,
		})
		return nil
	}
	s.store, _ = core.NewStore()
	s.notifies = [][]interface{}{}
	s.eng = core.NewEngine(s.store)
	s.eng.AddPlugin(s.plugin)
	s.eng.AddPlugin(NewLastFor(s.clock))
	s.eng.AddPlugin(s.notifyPlugin)
}

func (s *alertTestSuite) TestNormalToP0() {
	s.NoError(s.run(`
		alert(
			"cpu",
			notify,
			lastFor(10, ">", 5, "0m"),
			lastFor(true, "0m")
		)
	`))
	s.Equal(s.notifies, [][]interface{}{
		[]interface{}{
			Severity(0), false,
			"Incident test-job-id#cpu.P0 10.000000 > 5.000000",
		},
	})
}

func (s *alertTestSuite) TestNormalToP1() {
	s.NoError(s.run(`
		alert(
			"cpu",
			notify,
			lastFor(5, ">", 10, "0m"),
			lastFor(5, ">", 3, "0m")
		)
	`))
	s.Equal(s.notifies, [][]interface{}{
		[]interface{}{
			Severity(1), false,
			"Incident test-job-id#cpu.P1 5.000000 > 3.000000",
		},
	})
}

func (s *alertTestSuite) TestP1ToP0() {
	s.store.Set(alertNS, "test-job-id#cpu", &alertStatus{
		Severity: Severity(1),
	})
	s.NoError(s.run(`
		alert(
			"cpu",
			notify,
			lastFor(10, ">", 5, "0m"),
			lastFor(5, ">", 3, "0m")
		)
	`))
	s.Equal(s.notifies, [][]interface{}{
		[]interface{}{
			Severity(0), false,
			"Incident test-job-id#cpu.P0 10.000000 > 5.000000",
		},
	})
}

func (s *alertTestSuite) TestP1ToP2() {
	s.store.Set(alertNS, "test-job-id#cpu", &alertStatus{
		Severity: Severity(1),
	})
	s.NoError(s.run(`
		alert(
			"cpu",
			notify,
			lastFor(2, ">", 5, "0m"),
			lastFor(2, ">", 3, "0m"),
			lastFor(2, ">", 1, "0m")
		)
	`))
	s.Equal(s.notifies, [][]interface{}{
		[]interface{}{
			Severity(1), true,
			"Resolve test-job-id#cpu.P1 2.000000 > 3.000000",
		},
		[]interface{}{
			Severity(2), false,
			"Incident test-job-id#cpu.P2 2.000000 > 1.000000",
		},
	})
}

func (s *alertTestSuite) TestP1ToNormal() {
	s.store.Set(alertNS, "test-job-id#cpu", &alertStatus{
		Severity: Severity(1),
	})
	s.NoError(s.run(`
		alert(
			"cpu",
			notify,
			lastFor(0, ">", 5, "0m"),
			lastFor(0, ">", 3, "0m"),
			lastFor(0, ">", 1, "0m")
		)
	`))
	s.Equal(s.notifies, [][]interface{}{
		[]interface{}{
			Severity(1), true,
			"Resolve test-job-id#cpu.P1 0.000000 > 3.000000",
		},
	})
}

func (s *alertTestSuite) run(script string) error {
	meta := sauron.JobMeta{
		JobID:  testJobID,
		Script: script,
	}
	s.eng.SetJobMeta(meta)
	return s.eng.Run()
}
