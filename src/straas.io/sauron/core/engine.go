package core

import (
	"fmt"

	"github.com/robertkrimen/otto"

	"straas.io/sauron"
)

const (
	dummyStatement = `1==1;`
)

// NewEngine creates an instance of engine
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
	e.vm.Set(p.Name(), e.makeOttoFunc(p.Run))
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
	// add one more dummy statement for vm to have chance to throw exception
	script := fmt.Sprintf("%s\n%s", e.meta.Script, dummyStatement)

	_, err = e.vm.Run(script)
	return err
}

// makeOttoFunc wrap plugin by otto
func (e *engineImpl) makeOttoFunc(run func(sauron.PluginContext) error) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		// prepare context
		ctx := &contextImpl{
			call: call,
			eng:  e,
		}

		// terminate VM if error occurs
		if err := run(ctx); err != nil {
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
	eng      *engineImpl
	store    sauron.Store
	call     otto.FunctionCall
	rtnValue interface{}
}

func (c *contextImpl) JobMeta() sauron.JobMeta {
	return c.eng.meta
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

func (c *contextImpl) CallFunction(i int, args ...interface{}) (interface{}, error) {
	arg, err := c.getArg(i)
	if err != nil {
		return nil, err
	}
	if !arg.IsFunction() {
		return nil, fmt.Errorf("arg %d is not a function", i)
	}
	this := otto.Value{}
	result, err := arg.Call(this, args...)
	if err != nil {
		return nil, err
	}
	return result.Export()
}

func (c *contextImpl) ArgLen() int {
	return len(c.call.ArgumentList)
}

func (c *contextImpl) Return(v interface{}) error {
	// TODO: check invlid return
	// convert to otto func
	if funcRtn, ok := v.(sauron.FuncReturn); ok {
		c.rtnValue = c.eng.makeOttoFunc(funcRtn)
		return nil
	}
	c.rtnValue = v
	return nil
}

func (c *contextImpl) Store() sauron.Store {
	return c.store
}

func (c *contextImpl) getArg(i int) (otto.Value, error) {
	if i < 0 || i >= c.ArgLen() {
		return otto.Value{}, fmt.Errorf("index %d out of bound", i)
	}
	return c.call.Argument(i), nil
}
