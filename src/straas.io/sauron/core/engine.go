package core

import (
	"fmt"

	"github.com/robertkrimen/otto"

	"straas.io/sauron"
)

func NewEngine() sauron.Engine {
	return &engineImpl{
		vm: otto.New(),
	}
}

type engineImpl struct {
	vm   *otto.Otto
	meta sauron.JobMeta
}

func (e *engineImpl) SetJobMeta(meta sauron.JobMeta) error {
	e.meta = meta
	return nil
}

// register plugin funcs
func (e *engineImpl) AddPlugin(p sauron.Plugin) error {
	e.vm.Set(p.Name(), e.makeOttoFunc(p))
	return nil
}

func (e *engineImpl) Run() (err error) {
	defer func() {
		if caught := recover(); caught != nil {
			err = fmt.Errorf("fail to run, err:%v", caught)
			return
		}
	}()
	// buffered to avoid blocking
	e.vm.Interrupt = make(chan func(), 1)
	// TODO: cache compiled script for better performance
	_, err = e.vm.Run(e.meta.Script)
	return err
}

// makeOttoFunc wrap plugin by otto
func (e *engineImpl) makeOttoFunc(p sauron.Plugin) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		// prepare context
		ctx := &contextImpl{
			meta: e.meta,
			call: call,
		}

		// terminate VM if error occurs
		if err := p.Run(ctx); err != nil {
			// halt
			e.haltVM(err)
			return otto.Value{}
		}

		result, err := e.vm.ToValue(ctx.rtnValue)
		if err != nil {
			// halt
			e.haltVM(err)
			return otto.Value{}
		}
		return result
	}
}

func (e *engineImpl) haltVM(err error) {
	// refer https://github.com/robertkrimen/otto#halting-problem
	e.vm.Interrupt <- func() {
		panic(err)
	}
	// TODO: need to close Interrupt channel ?!
	// close(e.vm.Interrupt)
}

// contextImpl implements Context
type contextImpl struct {
	meta     sauron.JobMeta
	call     otto.FunctionCall
	rtnValue interface{}
}

func (c *contextImpl) JobMeta() sauron.JobMeta {
	return c.meta
}

func (c *contextImpl) ArgBoolean(i int) (bool, error) {
	arg, err := c.getArg(i)
	if err != nil {
		return false, err
	}
	if !arg.IsBoolean() {
		return false, fmt.Errorf("arg %d is not a boolean", i)
	}
	return arg.ToBoolean()
}

func (c *contextImpl) ArgInt(i int) (int64, error) {
	arg, err := c.getArg(i)
	if err != nil {
		return 0, err
	}
	if !arg.IsNumber() {
		return 0, fmt.Errorf("arg %d is not a number", i)
	}
	return arg.ToInteger()
}

func (c *contextImpl) ArgFloat(i int) (float64, error) {
	arg, err := c.getArg(i)
	if err != nil {
		return 0, err
	}
	if !arg.IsNumber() {
		return 0, fmt.Errorf("arg %d is not a number", i)
	}
	return arg.ToFloat()
}

func (c *contextImpl) ArgString(i int) (string, error) {
	arg, err := c.getArg(i)
	if err != nil {
		return "", err
	}
	if !arg.IsString() {
		return "", fmt.Errorf("arg %d is not a number", i)
	}
	return arg.ToString()
}

func (c *contextImpl) ArgLen() int {
	return len(c.call.ArgumentList)
}

func (c *contextImpl) Return(v interface{}) error {
	// TODO: check invlid return
	c.rtnValue = v
	return nil
}

func (c *contextImpl) getArg(i int) (otto.Value, error) {
	if i < 0 || i >= c.ArgLen() {
		return otto.Value{}, fmt.Errorf("index %d out of bound", i)
	}
	return c.call.Argument(i), nil
}
