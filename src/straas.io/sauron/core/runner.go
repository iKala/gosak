package core

import (
	"time"

	"straas.io/base/logger"
	"straas.io/base/timeutil"
	"straas.io/sauron"
)

const (
	// store namespace
	nsJobRunner    = "job-runner"
	schedErrTime   = 3 * time.Second
	chanBufferSize = 100
	maxConsErrs    = 10
)

var (
	log = logger.Get()
)

// NewJobRunner creates a job runner
func NewJobRunner(
	runnerID string,
	engFactory sauron.EngineFactory,
	ticker <-chan time.Time,
	store sauron.Store,
	jobs []sauron.JobMeta,
	plugins []sauron.Plugin,
	clock timeutil.Clock,
) sauron.JobRunner {
	return &jobRunner{
		runnerID:   runnerID,
		engFactory: engFactory,
		jobs:       jobs,
		plugins:    plugins,
		store:      store,
		ticker:     ticker,
		clock:      clock,

		events:  make(chan sauron.JobEvent, chanBufferSize),
		done:    make(chan bool, 1),
		update:  make(chan []sauron.JobMeta, 1),
		success: make(chan sauron.JobMeta, chanBufferSize),
		fail:    make(chan sauron.JobMeta, chanBufferSize),
	}
}

// jobRunner runner is responsible for invoking jobs according to their interval
// and record job status and result
// TODO: more logs
// TODO: record running count
// TODO: timeout mechanism
type jobRunner struct {
	runnerID    string
	engFactory  sauron.EngineFactory
	store       sauron.Store
	plugins     []sauron.Plugin
	jobs        []sauron.JobMeta
	runningJobs map[string]*jobStatus
	clock       timeutil.Clock

	events  chan sauron.JobEvent
	ticker  <-chan time.Time
	done    chan bool
	update  chan []sauron.JobMeta
	success chan sauron.JobMeta
	fail    chan sauron.JobMeta
}

// jobStatus records job status
// using unixts to avoid too larg json marshal problem
type jobStatus struct {
	// RunnerID is current job runner id
	RunnerID string
	// LastRun is last job start running time
	LastRun int64
	// LastSuccess is last success time
	LastSuccess int64
	// LastSuccess is last fail time
	LastFail int64
	// Running indicates whether job is running
	Running bool
	// ConsError is the number of consecutive errors
	ConsError int
}

// Start starts the job runner
func (j *jobRunner) Start() error {
	go j.runLoop()
	return nil
}

// Stop stops the job runner
func (j *jobRunner) Stop() error {
	j.done <- true
	return nil
}

// Update updates all jobs
func (j *jobRunner) Update(jobs []sauron.JobMeta) error {
	// update jobs
	j.update <- jobs
	return nil
}

func (j *jobRunner) Events() <-chan sauron.JobEvent {
	return j.events
}

// insertJobs insert all jobs to queue
func (j *jobRunner) insertJobs(jobs []sauron.JobMeta) {
	for _, meta := range jobs {
		j.updateJobStatus(meta.JobID, func(status *jobStatus) (*jobStatus, error) {
			now := j.clock.Now()
			lastRun := time.Unix(status.LastRun, 0)
			// still try to run if not the same runner
			if status.RunnerID == j.runnerID && status.Running {
				log.Infof("job %s is alreay running", meta.JobID)
				// report timeout, cannot finish in a running interval
				if now.Sub(lastRun) > meta.Interval {
					j.sendEvent(meta.JobID, sauron.EventTimeout)
				}
				return nil, nil
			}

			// check running interval with allowed error
			if now.Sub(lastRun)+schedErrTime < meta.Interval {
				log.Infof("job %s waits for next round", meta.JobID)
				return nil, nil
			}
			if err := j.invokeJob(meta); err != nil {
				log.Infof("fail to invoke job %s, err:%v", err)
				return nil, err
			}
			log.Infof("job %s invoked", meta.JobID)
			status.Running = true
			status.LastRun = now.Unix()
			return status, nil
		})
	}
}

func (j *jobRunner) runLoop() {
	log.Infof("start job runner loop")
	for j.loopOnce() {
	}
}

// use event loop to avoid status racing condition
// loopOnce is for testing purpose, return value indicates
// continue
func (j *jobRunner) loopOnce() bool {
	select {
	// scheduler terminated
	case <-j.done:
		return false

	case <-j.ticker:
		log.Infof("insert jobs %d", len(j.jobs))
		j.insertJobs(j.jobs)

	case jobs := <-j.update:
		log.Infof("update jobs %d", len(jobs))
		j.jobs = jobs

	// job success
	case meta := <-j.success:
		j.updateJobStatus(meta.JobID, func(status *jobStatus) (*jobStatus, error) {
			status.Running = false
			status.LastSuccess = j.clock.Now().Unix()
			status.ConsError = 0
			return status, nil
		})

	// job fail
	case meta := <-j.fail:
		j.updateJobStatus(meta.JobID, func(status *jobStatus) (*jobStatus, error) {
			status.Running = false
			status.LastFail = j.clock.Now().Unix()
			status.ConsError++
			// report consecutive errors
			if status.ConsError >= maxConsErrs {
				j.sendEvent(meta.JobID, sauron.EventConsErr)
			}
			return status, nil
		})
	}
	return true
}

func (j *jobRunner) invokeJob(meta sauron.JobMeta) error {
	eng := j.engFactory()
	if err := eng.SetJobMeta(meta); err != nil {
		return err
	}
	for _, p := range j.plugins {
		if err := eng.AddPlugin(p); err != nil {
			return err
		}
	}
	// invoke a goroutine to run jobs parallelly
	go func() {
		if err := eng.Run(); err != nil {
			log.Errorf("execute %s fail, err:%v", meta.JobID, err)
			j.fail <- meta
			return
		}
		log.Infof("execute %s success", meta.JobID)
		j.success <- meta
	}()
	return nil
}

func (j *jobRunner) sendEvent(jobID string, code sauron.EventCode) {
	log.Infof("send event jobID:%s, code:%v", jobID, code)
	j.events <- sauron.JobEvent{
		JobID: jobID,
		Code:  code,
	}
}

// updateJobStatus get jobStatus by jobId, process by action and then write back to store
func (j *jobRunner) updateJobStatus(jobID string, action func(status *jobStatus) (*jobStatus, error)) error {
	status, err := j.ensureJobStatus(jobID)
	if err != nil {
		return err
	}
	status, err = action(status)
	if err != nil {
		return err
	}
	// ignore update
	if status == nil {
		return nil
	}
	status.RunnerID = j.runnerID
	return j.setJobStatus(jobID, status)
}

// ensureJobStatus gets jobStatus from store or creates one if does not exists
func (j *jobRunner) ensureJobStatus(jobID string) (*jobStatus, error) {
	status := &jobStatus{}
	ok, err := j.store.Get(nsJobRunner, jobID, status)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &jobStatus{
			RunnerID: j.runnerID,
		}, nil
	}
	return status, nil
}

// setJobStatus saves jobStatus back to store
func (j *jobRunner) setJobStatus(jobID string, status *jobStatus) error {
	return j.store.Set(nsJobRunner, jobID, status)
}
