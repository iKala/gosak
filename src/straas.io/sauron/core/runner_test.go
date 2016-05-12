package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

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
		curTime: time.Now(),
		store:   store,
	}
	timeNow = func() time.Time {
		return c.curTime
	}
	engFactory := func() sauron.Engine {
		return c.engFactory()
	}
	plugins := []sauron.Plugin{c.plugin1, c.plugin2}
	c.runner = NewJobRunner(testRunnerID, engFactory, c.ticker, c.store, nil, plugins).(*jobRunner)
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

func (s *runnerTestSuite) TestJobStatusStore() {
	c := newContext(s.T())
	testJobID := "test-job-id"
	testStatus := &jobStatus{
		Running: false,
		LastRun: 1234,
	}

	// test no exists jobs
	status, err := c.runner.ensureJobStatus(testJobID)
	// got an empty job
	s.NoError(err)
	s.Equal(status, &jobStatus{RunnerID: testRunnerID})

	// test set then get
	c.runner.setJobStatus(testJobID, testStatus)
	status, err = c.runner.ensureJobStatus(testJobID)
	s.NoError(err)
	s.Equal(status, testStatus)
}

func (s *runnerTestSuite) TestUpdateJobStatus() {
	c := newContext(s.T())
	testJobID := "test-job-id"
	testStatus := &jobStatus{
		Running: false,
		LastRun: 1234,
	}
	testChgedStatus := &jobStatus{
		Running:  true,
		LastRun:  789,
		RunnerID: testRunnerID,
	}

	called := false
	c.runner.setJobStatus(testJobID, testStatus)
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
	status, _ := c.runner.ensureJobStatus(testJobID)
	s.Equal(status, testChgedStatus)
}

func (s *runnerTestSuite) TestUpdateJobStatusNotUpdate() {
	c := newContext(s.T())
	testJobID := "test-job-id"
	testStatus := &jobStatus{
		Running: false,
		LastRun: 1234,
	}

	called := false
	c.runner.setJobStatus(testJobID, testStatus)
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
	status, _ := c.runner.ensureJobStatus(testJobID)
	s.Equal(status, testStatus)
}

func (s *runnerTestSuite) TestUpdateJobStatusError() {
	c := newContext(s.T())
	testJobID := "test-job-id"
	testStatus := &jobStatus{
		Running: false,
		LastRun: 1234,
	}

	called := false
	c.runner.setJobStatus(testJobID, testStatus)
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
	status, _ := c.runner.ensureJobStatus(testJobID)
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
	c.ticker <- timeNow()
	c.loopOnceTimeout()

	// 3 jobs invoked
	// all should be running
	for _, job := range testJobs {
		status, _ := c.runner.ensureJobStatus(job.JobID)
		s.Equal(status.LastRun, c.curTime.Unix())
		s.True(status.Running)
	}

	// tick
	c.curTime = c.curTime.Add(time.Minute)

	// 3 jobs report
	c.loopOnceTimeout()
	c.loopOnceTimeout()
	c.loopOnceTimeout()

	for _, eng := range engs {
		eng.AssertExpectations(s.T())
	}

	// all should not running
	for _, job := range testJobs {
		status, _ := c.runner.ensureJobStatus(job.JobID)
		s.Equal(status.LastSuccess, c.curTime.Unix())
		s.False(status.Running)
	}
}

func (s *runnerTestSuite) TestIgnoreRunJob() {
	c := newContext(s.T())
	c.runner.Update(testJobs)
	c.loopOnceTimeout()

	c.runner.setJobStatus(testJobs[0].JobID, &jobStatus{
		RunnerID: testRunnerID,
		Running:  true,
	})
	c.runner.setJobStatus(testJobs[1].JobID, &jobStatus{
		LastRun: timeNow().Unix(),
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
	c.ticker <- timeNow()
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
	c.ticker <- timeNow()
	c.loopOnceTimeout()

	// tick
	c.curTime = c.curTime.Add(time.Minute)

	// one report
	c.loopOnceTimeout()
	c.loopOnceTimeout()
	c.loopOnceTimeout()

	for _, eng := range engs {
		eng.AssertExpectations(s.T())
	}

	// all should not running
	for _, job := range testJobs {
		status, _ := c.runner.ensureJobStatus(job.JobID)
		s.False(status.Running)
		if job.JobID == testJobs[2].JobID {
			s.Equal(status.LastFail, c.curTime.Unix())
		} else {
			s.Equal(status.LastSuccess, c.curTime.Unix())
		}
	}
}
