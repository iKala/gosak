package core

import (
	"fmt"

	"github.com/fatih/color"

	"straas.io/sauron"
)

// NewOutput creates a output instance
func NewOutput(dryRun bool) sauron.Output {
	return &outputImpl{
		dryRun: dryRun,
	}
}

type outputImpl struct {
	dryRun bool
}

func (o *outputImpl) Infof(format string, args ...interface{}) {
	if o.dryRun {
		color.Green(fmt.Sprintf(format, args...))
	}
	log.Infof(format, args...)

}

func (o *outputImpl) Errorf(format string, args ...interface{}) {
	if o.dryRun {
		color.Red(fmt.Sprintf(format, args...))
	}
	log.Errorf(format, args...)
}
