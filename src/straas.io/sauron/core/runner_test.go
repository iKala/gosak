package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"straas.io/base/timeutil"
	"straas.io/sauron"
	"straas.io/sauron/mocks"
)

var (
	testJobs = []sauron.JobMeta{
		sauron.JobMeta{
			JobID:    "job1",
			Interval: time.Minute,
		},
		sauron.JobMeta{
			JobID:    "job2",
			Interval: time.Minute,
		},
		sauron.JobMeta{
			JobID:    "job3",
			Interval: time.Minute,
		},
	}
	testRunnerID = "aaaa"
)

// TODO: more test
func TestRunnerSuite(t *testing.T) {
	suite.Run(t, new(runnerTestSuite))
}

type runnerTestSuite struct {
	suite.Suite
}

func newContext(t *testing.T) *testContext {
	store, _ := NewStore()
	c := &testContext{
		t:       t,
		plugin1: &mocks.Plugin{},
		plugin2: &mocks.Plugin{},
		ticker:  make(chan time.Time, 1),
		store:   store,
		clock:   timeutil.NewFakeClock(),
	}
	engFactory := func() sauron.Engine {
		return c.engFactory()
	}
	plugins := []sauron.Plugin{c.plugin1, c.plugin2}
	c.runner = NewJobRunner(testRunnerID, engFactory, c.ticker,
		c.store, nil, plugins, c.clock).(*jobRunner)
	return c
}

type testContext struct {
	t          *testing.T
	runner     *jobRunner
	ticker     chan time.Time
	plugin1    *mocks.Plugin
	plugin2    *mocks.Plugin
	engFactory sauron.EngineFactory
	store      sauron.Store
	curTime    time.Time
	clock      timeutil.FakeClock
}

func (c *testContext) loopOnceTimeout() {
	done := make(chan bool)
	go func() {
		c.runner.loopOnce()
		close(done)
	}()
	select {
	case <-time.After(5 * time.Second):
		c.t.Error("execute timeout")
	case <-done:
	}
}

func (s *runnerTestSuite) TestUpdateJobStatus() {
	c := newContext(s.T())
	testJobID := "test-job-id"
	testStatus := &jobStatus{
		Running:  false,
		LastRun:  1234,
		RunnerID: testRunnerID,
	}
	testChgedStatus := &jobStatus{
		Running:  true,
		LastRun:  789,
		RunnerID: testRunnerID,
	}

	called := false
	setStatus(c.store, testJobID, testStatus)
	err := c.runner.updateJobStatus(testJobID, func(status *jobStatus) (*jobStatus, error) {
		called = true
		s.Equal(status, testStatus)
		status.Running = true
		status.LastRun = 789
		return status, nil
	})
	// make sure really called
	s.True(called)
	s.NoError(err)

	// get changed status
	status := getStatus(c.store, testJobID)
	s.Equal(status, testChgedStatus)
}

func (s *runnerTestSuite) TestUpdateJobStatusNotUpdate() {
	c := newContext(s.T())
	testJobID := "test-job-id"
	testStatus := &jobStatus{
		RunnerID: testRunnerID,
		Running:  false,
		LastRun:  1234,
	}

	called := false
	setStatus(c.store, testJobID, testStatus)
	err := c.runner.updateJobStatus(testJobID, func(status *jobStatus) (*jobStatus, error) {
		called = true
		s.Equal(status, testStatus)
		status.Running = true
		status.LastRun = 789
		// return nil indicate not to update
		return nil, nil
	})
	// make sure really called
	s.True(called)
	s.NoError(err)

	// get changed status
	status := getStatus(c.store, testJobID)
	s.Equal(status, testStatus)
}

func (s *runnerTestSuite) TestUpdateJobStatusError() {
	c := newContext(s.T())
	testJobID := "test-job-id"
	testStatus := &jobStatus{
		Running:  false,
		LastRun:  1234,
		RunnerID: testRunnerID,
	}

	called := false
	setStatus(c.store, testJobID, testStatus)
	err := c.runner.updateJobStatus(testJobID, func(status *jobStatus) (*jobStatus, error) {
		called = true
		s.Equal(status, testStatus)
		status.Running = true
		status.LastRun = 789
		// return nil indicate not to update
		return nil, fmt.Errorf("some err")
	})
	// make sure really called
	s.True(called)
	s.Error(err)

	// get changed status
	status := getStatus(c.store, testJobID)
	s.Equal(status, testStatus)
}

func (s *runnerTestSuite) TestUpdateJobs() {
	c := newContext(s.T())
	c.runner.Update(testJobs)
	c.loopOnceTimeout()
	s.Equal(c.runner.jobs, testJobs)
}

func (s *runnerTestSuite) TestRunJob() {
	c := newContext(s.T())
	c.runner.Update(testJobs)
	c.loopOnceTimeout()

	// is empty
	engs := []*mocks.Engine{}
	jobIdx := 0

	c.engFactory = func() sauron.Engine {
		eng := &mocks.Engine{}
		eng.On("SetJobMeta", testJobs[jobIdx]).Return(nil).Once()
		eng.On("AddPlugin", c.plugin1).Return(nil).Once()
		eng.On("AddPlugin", c.plugin2).Return(nil).Once()
		eng.On("Run").Return(nil).Once()

		jobIdx++
		engs = append(engs, eng)
		return eng
	}

	// loop once, should add jobs to pending
	c.ticker <- c.clock.Now()
	curTime := c.clock.Now()
	c.loopOnceTimeout()

	// 3 jobs invoked
	// all should be running
	for _, job := range testJobs {
		status := getStatus(c.store, job.JobID)
		s.Equal(status.LastRun, curTime.Unix())
		s.True(status.Running)
	}

	// tick
	curTime = c.clock.Now()

	// 3 jobs report
	c.loopOnceTimeout()
	c.loopOnceTimeout()
	c.loopOnceTimeout()

	for _, eng := range engs {
		eng.AssertExpectations(s.T())
	}

	// all should not running
	for _, job := range testJobs {
		status := getStatus(c.store, job.JobID)
		s.Equal(status.LastSuccess, curTime.Unix())
		s.False(status.Running)
	}
}

func (s *runnerTestSuite) TestIgnoreRunJob() {
	c := newContext(s.T())
	c.runner.Update(testJobs)
	c.loopOnceTimeout()

	setStatus(c.store, testJobs[0].JobID, &jobStatus{
		RunnerID: testRunnerID,
		Running:  true,
	})
	setStatus(c.store, testJobs[1].JobID, &jobStatus{
		LastRun: c.clock.Now().Unix(),
	})

	engs := []*mocks.Engine{}
	c.engFactory = func() sauron.Engine {
		eng := &mocks.Engine{}
		eng.On("SetJobMeta", testJobs[2]).Return(nil).Once()
		eng.On("AddPlugin", c.plugin1).Return(nil).Once()
		eng.On("AddPlugin", c.plugin2).Return(nil).Once()
		eng.On("Run").Return(nil).Once()

		engs = append(engs, eng)
		return eng
	}

	// 3 jobs, but only one invoked
	c.ticker <- c.clock.Now()
	c.loopOnceTimeout()
	// one report
	c.loopOnceTimeout()

	for _, eng := range engs {
		eng.AssertExpectations(s.T())
	}
}

func (s *runnerTestSuite) TestRunJobError() {
	c := newContext(s.T())
	c.runner.Update(testJobs)
	c.loopOnceTimeout()

	engs := []*mocks.Engine{}
	jobIdx := 0

	c.engFactory = func() sauron.Engine {
		eng := &mocks.Engine{}
		eng.On("SetJobMeta", testJobs[jobIdx]).Return(nil).Once()
		eng.On("AddPlugin", c.plugin1).Return(nil).Once()
		eng.On("AddPlugin", c.plugin2).Return(nil).Once()

		if jobIdx == 2 {
			eng.On("Run").Return(fmt.Errorf("some error")).Once()
		} else {
			eng.On("Run").Return(nil).Once()
		}

		jobIdx++
		engs = append(engs, eng)
		return eng
	}

	// 3 jobs
	c.ticker <- c.clock.Now()
	c.loopOnceTimeout()

	// tick
	curTime := c.clock.Incr(time.Minute)

	// 3 jobs report
	c.loopOnceTimeout()
	c.loopOnceTimeout()
	c.loopOnceTimeout()

	for _, eng := range engs {
		eng.AssertExpectations(s.T())
	}

	// all should not running
	for _, job := range testJobs {
		status := getStatus(c.store, job.JobID)
		s.False(status.Running)
		if job.JobID == testJobs[2].JobID {
			s.Equal(status.LastFail, curTime.Unix())
		} else {
			s.Equal(status.LastSuccess, curTime.Unix())
		}
	}
}

func setStatus(store sauron.Store, jobID string, status *jobStatus) {
	store.Set(nsJobRunner, jobID, status)
}

func getStatus(store sauron.Store, jobID string) *jobStatus {
	v := &jobStatus{}
	ok, _ := store.Get(nsJobRunner, jobID, v)
	if ok {
		return v
	}
	return &jobStatus{}
}
