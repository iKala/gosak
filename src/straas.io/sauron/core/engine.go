package core

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/robertkrimen/otto"

	"straas.io/sauron"
)

const (
	dummyStatement = `1==1;`
)

var (
	emptyValue = reflect.Value{}
)

// NewEngine creates an instance of engine
func NewEngine(store sauron.Store, output sauron.Output) sauron.Engine {
	return &engineImpl{
		vm:    otto.New(),
		store: store,
	}
}

type engineImpl struct {
	vm     *otto.Otto
	store  sauron.Store
	output sauron.Output
	meta   sauron.JobMeta
	// using atomic for swap
	halt int64
}

func (e *engineImpl) SetJobMeta(meta sauron.JobMeta) error {
	e.meta = meta
	return nil
}

// register plugin funcs
func (e *engineImpl) AddPlugin(p sauron.Plugin) error {
	e.vm.Set(p.Name(), e.makeOttoFunc(p.Name(), p.Run))
	return nil
}

func (e *engineImpl) Run() (err error) {
	defer func() {
		if caught := recover(); caught != nil {
			err = fmt.Errorf("fail to run, err:%v", caught)
			return
		}
	}()
	log.Infof("[engine] run job, id:%s, env:%s, dryRun:%v",
		e.meta.JobID, e.meta.Env, e.meta.DryRun)

	// buffered to avoid blocking
	e.vm.Interrupt = make(chan func(), 1)
	// TODO: cache compiled script for better performance
	// add one more dummy statement for vm to have chance to throw exception
	script := fmt.Sprintf("%s\n%s", e.meta.Script, dummyStatement)

	_, err = e.vm.Run(script)
	return err
}

func (e *engineImpl) Get(name string) (interface{}, error) {
	v, err := e.vm.Get(name)
	if err != nil {
		return nil, err
	}
	return v.Export()
}

func (e *engineImpl) Set(name string, v interface{}) error {
	value, err := e.vm.ToValue(v)
	if err != nil {
		return err
	}
	return e.vm.Set(name, value)
}

// makeOttoFunc wrap plugin by otto
func (e *engineImpl) makeOttoFunc(name string,
	run func(sauron.PluginContext) error) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {

		log.Debug("[engine] ", callToString(name, call))

		// prepare context
		ctx := &contextImpl{
			call:   call,
			eng:    e,
			store:  e.store,
			Output: e.output,
		}

		// terminate VM if error occurs
		if err := run(ctx); err != nil {
			// halt
			e.haltVM(fmt.Errorf("[%s] %v", name, err))
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
	go func() {
		e.vm.Interrupt <- func() {
			panic(err)
		}
	}()
	// TODO: need to close Interrupt channel ?!
	// close(e.vm.Interrupt)
}

// callToString dumps call arguments as string for logging
func callToString(name string, call otto.FunctionCall) string {
	argStr := make([]string, 0, len(call.ArgumentList))
	for _, v := range call.ArgumentList {
		if v.IsFunction() {
			argStr = append(argStr, "<func>")
			continue
		}
		argStr = append(argStr, v.String())
	}

	return fmt.Sprintf("call %s(%s)", name, strings.Join(argStr, ", "))

}

// contextImpl implements Context
type contextImpl struct {
	eng      *engineImpl
	store    sauron.Store
	call     otto.FunctionCall
	rtnValue interface{}
	sauron.Output
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

func (c *contextImpl) ArgObject(i int, v interface{}) error {
	arg, err := c.getArg(i)
	if err != nil {
		return err
	}
	if !arg.IsObject() {
		return fmt.Errorf("arg %d is not an object", i)
	}
	vm, err := arg.Export()
	if err != nil {
		return err
	}
	return valueToObject(vm, v)
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

func (c *contextImpl) IsCallable(i int) bool {
	arg, err := c.getArg(i)
	if err != nil {
		return false
	}
	return arg.IsFunction()
}

func (c *contextImpl) ArgLen() int {
	return len(c.call.ArgumentList)
}

func (c *contextImpl) Return(v interface{}) error {
	// TODO: check invlid return
	// convert to otto func
	if funcRtn, ok := v.(sauron.FuncReturn); ok {
		c.rtnValue = c.eng.makeOttoFunc("<anonymous>", funcRtn)
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

func valueToObject(vm interface{}, obj interface{}) error {
	m, ok := vm.(map[string]interface{})
	if !ok {
		return fmt.Errorf("fail to convert object to map, %+v", vm)
	}
	t := reflect.ValueOf(obj).Elem()
	for k, v := range m {
		val := t.FieldByName(k)
		if val == emptyValue {
			return fmt.Errorf("unable to find field %s", k)
		}
		val.Set(reflect.ValueOf(v))
	}
	return nil
}
