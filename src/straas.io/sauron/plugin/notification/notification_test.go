package notification

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	yaml "gopkg.in/yaml.v2"

	"straas.io/sauron"
	"straas.io/sauron/core"
	"straas.io/sauron/mocks"
	nmocks "straas.io/sauron/plugin/notification/mocks"
)

const (
	testSinkerType = "testSinker"
	testCfg        = `
groups:
  - name: backend
    sinkers:
      - type: testSinker
        severity: [0,1]
        recovery: [2]
      - type: testSinker
        severity: [1,2]
        recovery: [1]
`
)

func TestNotification(t *testing.T) {
	suite.Run(t, new(notificationTestSuite))
}

type notificationTestSuite struct {
	suite.Suite
	plugin      *notificationPlugin
	eng         sauron.Engine
	mockConfig  *mocks.Config
	mockSinkers []*nmocks.Sinker
}

type testSinkCfg struct {
	position int
}

func (s *notificationTestSuite) SetupTest() {
	s.plugin = &notificationPlugin{}
	store, _ := core.NewStore()
	s.eng = core.NewEngine(store)
	s.eng.AddPlugin(s.plugin)

	sinkerFac = map[string]SinkerFactory{}
	s.mockSinkers = []*nmocks.Sinker{}
	s.mockConfig = &mocks.Config{}

	// fake
	fakeImpl := func(path string, v interface{}) error {
		return yaml.Unmarshal([]byte(testCfg), v)
	}
	s.mockConfig.On("LoadConfig", "notification/config", mock.Anything).Return(fakeImpl)

	RegisterSinker(testSinkerType, func() Sinker {
		ms := &nmocks.Sinker{}
		s.mockSinkers = append(s.mockSinkers, ms)
		fac := &testSinkCfg{position: len(s.mockSinkers)}
		ms.On("ConfigFactory").Return(fac)
		return ms
	})
}

func (s *notificationTestSuite) TestContain() {
	svs := []sauron.Severity{1, 2}
	s.False(contains(svs, sauron.Severity(0)))
	s.True(contains(svs, sauron.Severity(1)))
	s.True(contains(svs, sauron.Severity(2)))
}

func (s *notificationTestSuite) TestNewSinker() {
	sinker, err := newSinker(testSinkerType)
	s.NoError(err)
	s.Equal(sinker, s.mockSinkers[0])

	_, err = newSinker("unknown")
	s.Error(err)
}

func (s *notificationTestSuite) TestSinkerGroup() {
	ms1, _ := newSinker(testSinkerType)
	ms2, _ := newSinker(testSinkerType)

	info := &groupInfo{
		baseCfgs: []*BaseSinkCfg{
			&BaseSinkCfg{
				Severity: []sauron.Severity{1, 2},
				Recovery: []sauron.Severity{1},
			},
			&BaseSinkCfg{
				Severity: []sauron.Severity{0, 2},
				Recovery: []sauron.Severity{1},
			},
		},
		sinkers:    []Sinker{ms1, ms2},
		sinkerCfgs: []interface{}{"aaa", "bbb"},
	}

	sg := newSinkerGroup(info, sauron.Severity(1), false)
	s.Equal(sg.cfgs, []interface{}{"aaa"})

	sg = newSinkerGroup(info, sauron.Severity(1), true)
	s.Equal(sg.cfgs, []interface{}{"aaa", "bbb"})
}

func (s *notificationTestSuite) TestLoadConfig() {
	s.NoError(s.plugin.loadConfig(s.mockConfig))
	s.Equal(len(s.mockSinkers), 2)

	ms1 := s.mockSinkers[0]
	ms2 := s.mockSinkers[1]
	expeced := map[string]*groupInfo{
		"backend": &groupInfo{
			baseCfgs: []*BaseSinkCfg{
				&BaseSinkCfg{
					Type:     testSinkerType,
					Severity: []sauron.Severity{0, 1},
					Recovery: []sauron.Severity{2},
				},
				&BaseSinkCfg{
					Type:     testSinkerType,
					Severity: []sauron.Severity{1, 2},
					Recovery: []sauron.Severity{1},
				},
			},
			sinkerCfgs: []interface{}{
				&testSinkCfg{position: int(1)},
				&testSinkCfg{position: int(2)},
			},
			sinkers: []Sinker{ms1, ms2},
		},
	}
	s.Equal(s.plugin.groupInfoMap, expeced)
}

func (s *notificationTestSuite) TestRunAlert() {
	s.NoError(s.plugin.loadConfig(s.mockConfig))
	s.Equal(len(s.mockSinkers), 2)

	ms1 := s.mockSinkers[0]
	ms2 := s.mockSinkers[1]
	cfg1 := &testSinkCfg{position: int(1)}
	cfg2 := &testSinkCfg{position: int(2)}

	ms1.On("Sink", cfg1, sauron.Severity(1), false, "some desc").Return(nil).Once()
	ms2.On("Sink", cfg2, sauron.Severity(1), false, "some desc").Return(nil).Once()
	s.run(`
		notify("backend")(1, false, "some desc")
	`)
	ms1.AssertExpectations(s.T())
	ms2.AssertExpectations(s.T())
}

func (s *notificationTestSuite) TestRun1Resolve() {
	s.NoError(s.plugin.loadConfig(s.mockConfig))
	s.Equal(len(s.mockSinkers), 2)

	ms1 := s.mockSinkers[0]
	cfg1 := &testSinkCfg{position: int(1)}

	ms1.On("Sink", cfg1, sauron.Severity(2), true, "some desc").Return(nil).Once()
	s.run(`
		notify("backend")(2, true, "some desc")
	`)
	ms1.AssertExpectations(s.T())
}

func (s *notificationTestSuite) run(script string) error {
	meta := sauron.JobMeta{
		Script: script,
	}
	s.eng.SetJobMeta(meta)
	return s.eng.Run()
}
