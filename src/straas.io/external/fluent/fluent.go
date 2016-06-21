package fluent

import (
	"fmt"
	"os"

	"github.com/fluent/fluent-logger-golang/fluent"

	"straas.io/external"
)

// New creates an instance of fluent client
func New(enable bool, host string, port int) (external.Fluent, error) {
	if !enable {
		return &dummyImpl{}, nil
	}

	logger, err := fluent.New(fluent.Config{
		FluentHost: host,
		FluentPort: port,
	})
	if err != nil {
		return nil, err
	}
	return &fluentImpl{
		logger: logger,
	}, nil
}

// TODO: old version logger is not thread-safe, need to
// check new version
type fluentImpl struct {
	logger *fluent.Fluent
}

func (f *fluentImpl) Post(tag string, v interface{}) {
	if err := f.logger.Post(tag, v); err != nil {
		// since fluent is for logs or metrics
		// once it has problems, printing logs or sending metrics is meaningless
		// so just print to console error
		fmt.Fprintf(os.Stderr, "fail to send fluent, err:%v\n", err)
	}
}

type dummyImpl struct {
}

func (f *dummyImpl) Post(tag string, v interface{}) {
	// nothing to do
}
