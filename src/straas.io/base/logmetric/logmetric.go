package logmetric

import (
	"github.com/facebookgo/stats"

	"straas.io/base/logger"
)

// New creates an instance of LogMetric
func New(log logger.Logger, stat stats.Client) LogMetric {
	return &logMetricImpl{
		Client: stat,
		Logger: log,
	}
}

// NewDummy creates a dummy instance of LogMetric
func NewDummy() LogMetric {
	return &dummyImpl{}
}

// LogMetric composites logger and metric
type LogMetric interface {
	stats.Client
	logger.Logger
}

type logMetricImpl struct {
	stats.Client
	logger.Logger
}

type dummyImpl struct {
}

func (*dummyImpl) BumpAvg(string, float64)       {}
func (*dummyImpl) BumpSum(string, float64)       {}
func (*dummyImpl) BumpHistogram(string, float64) {}
func (*dummyImpl) BumpTime(string) interface {
	End()
} {
	return stats.NoOpEnd
}

func (*dummyImpl) Debugf(string, ...interface{})   {}
func (*dummyImpl) Infof(string, ...interface{})    {}
func (*dummyImpl) Printf(string, ...interface{})   {}
func (*dummyImpl) Warnf(string, ...interface{})    {}
func (*dummyImpl) Warningf(string, ...interface{}) {}
func (*dummyImpl) Errorf(string, ...interface{})   {}
func (*dummyImpl) Fatalf(string, ...interface{})   {}
func (*dummyImpl) Panicf(string, ...interface{})   {}

func (*dummyImpl) Debug(...interface{})   {}
func (*dummyImpl) Info(...interface{})    {}
func (*dummyImpl) Print(...interface{})   {}
func (*dummyImpl) Warn(...interface{})    {}
func (*dummyImpl) Warning(...interface{}) {}
func (*dummyImpl) Error(...interface{})   {}
func (*dummyImpl) Fatal(...interface{})   {}
func (*dummyImpl) Panic(...interface{})   {}

func (*dummyImpl) Debugln(...interface{})   {}
func (*dummyImpl) Infoln(...interface{})    {}
func (*dummyImpl) Println(...interface{})   {}
func (*dummyImpl) Warnln(...interface{})    {}
func (*dummyImpl) Warningln(...interface{}) {}
func (*dummyImpl) Errorln(...interface{})   {}
func (*dummyImpl) Fatalln(...interface{})   {}
func (*dummyImpl) Panicln(...interface{})   {}
