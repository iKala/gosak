package core

import (
	"fmt"
	"testing"

	"github.com/robertkrimen/otto"
	"github.com/stretchr/testify/suite"

	"straas.io/sauron"
)

func TestEngineSuite(t *testing.T) {
	suite.Run(t, new(engineTestSuite))
}

func TestContentSuite(t *testing.T) {
	suite.Run(t, new(contextTestSuite))
}

type engineTestSuite struct {
	suite.Suite
	eng sauron.Engine
}

type testPlugin struct {
	name string
	run  func(ctx sauron.PluginContext) error
}

func (t *testPlugin) Name() string {
	return t.name
}

func (t *testPlugin) Run(ctx sauron.PluginContext) error {
	return t.run(ctx)
}

func (t *testPlugin) HelpMsg() string {
	return ""
}

func (s *engineTestSuite) SetupTest() {
	store, _ := NewStore()
	s.eng = NewEngine(store)
}

func (s *engineTestSuite) TestPlugins() {
	v := ""
	meta := sauron.JobMeta{
		Script: `bbb(aaa())`,
	}

	p1 := &testPlugin{
		name: "aaa",
		run: func(ctx sauron.PluginContext) error {
			s.Equal(meta, ctx.JobMeta())
			ctx.Return("abcd")
			return nil
		},
	}
	p2 := &testPlugin{
		name: "bbb",
		run: func(ctx sauron.PluginContext) error {
			s.Equal(meta, ctx.JobMeta())
			v, _ = ctx.ArgString(0)
			return nil
		},
	}

	err := s.eng.AddPlugin(p1)
	s.NoError(err)
	err = s.eng.AddPlugin(p2)
	s.NoError(err)

	err = s.eng.SetJobMeta(meta)
	s.NoError(err)
	s.NoError(s.eng.Run())
	s.Equal(v, "abcd")
}

func (s *engineTestSuite) TestPluginError() {
	p1Called := false
	p2Called := false
	p1 := &testPlugin{
		name: "aaa",
		run: func(ctx sauron.PluginContext) error {
			p1Called = true
			return fmt.Errorf("some error")
		},
	}
	p2 := &testPlugin{
		name: "bbb",
		run: func(ctx sauron.PluginContext) error {
			p2Called = true
			return nil
		},
	}

	err := s.eng.AddPlugin(p1)
	s.NoError(err)
	err = s.eng.AddPlugin(p2)
	s.NoError(err)

	err = s.eng.SetJobMeta(sauron.JobMeta{
		Script: `aaa(); bbb();`,
	})
	s.NoError(err)
	s.Error(s.eng.Run())
	s.True(p1Called)
	s.False(p2Called)
}

func (s *engineTestSuite) TestIllegalScript() {
	err := s.eng.SetJobMeta(sauron.JobMeta{
		Script: `aaa(;`,
	})
	s.NoError(err)
	s.Error(s.eng.Run())
}

// -----------------------------------------------------------------

type contextTestSuite struct {
	suite.Suite
	eng    *engineImpl
	called bool
	check  func(ctx *contextImpl)
}

func (s *contextTestSuite) SetupTest() {
	s.called = false
	p := &testPlugin{
		name: "test",
		run: func(ctx sauron.PluginContext) error {
			s.called = true
			s.check(ctx.(*contextImpl))
			return nil
		},
	}
	store, _ := NewStore()
	s.eng = NewEngine(store).(*engineImpl)
	s.eng.AddPlugin(p)
}

func (s *contextTestSuite) TestBool() {
	s.check = func(ctx *contextImpl) {
		b, err := ctx.ArgBoolean(0)
		s.True(b)
		s.NoError(err)

		b, err = ctx.ArgBoolean(1)
		s.False(b)
		s.NoError(err)

		_, err = ctx.ArgBoolean(2)
		s.Error(err)

		_, err = ctx.ArgBoolean(3)
		s.Error(err)
	}
	s.run(`test(true, false, "xxx")`)
}

func (s *contextTestSuite) TestInt() {
	s.check = func(ctx *contextImpl) {
		b, err := ctx.ArgInt(0)
		s.Equal(b, int64(10))
		s.NoError(err)

		b, err = ctx.ArgInt(1)
		s.Equal(b, int64(-3))
		s.NoError(err)

		b, err = ctx.ArgInt(2)
		s.Equal(b, int64(5))
		s.NoError(err)

		_, err = ctx.ArgInt(3)
		s.Error(err)

		_, err = ctx.ArgInt(4)
		s.Error(err)
	}
	s.run(`test(10, -3, 5.5, "xxx")`)
}

func (s *contextTestSuite) TestFloat() {
	s.check = func(ctx *contextImpl) {
		b, err := ctx.ArgFloat(0)
		s.Equal(b, float64(3.3))
		s.NoError(err)

		b, err = ctx.ArgFloat(1)
		s.Equal(b, float64(-4))
		s.NoError(err)

		_, err = ctx.ArgFloat(2)
		s.Error(err)

		_, err = ctx.ArgFloat(3)
		s.Error(err)
	}
	s.run(`test(3.3, -4, "xxx")`)
}

func (s *contextTestSuite) TestString() {
	s.check = func(ctx *contextImpl) {
		b, err := ctx.ArgString(0)
		s.Equal(b, "xxxx")
		s.NoError(err)

		b, err = ctx.ArgString(1)
		s.Equal(b, "yyy")
		s.NoError(err)

		_, err = ctx.ArgString(2)
		s.Error(err)

		_, err = ctx.ArgString(3)
		s.Error(err)
	}
	s.run(`test("xxxx", 'yyy', 33)`)
}

type testStruct struct {
	MyName string
	MyAge  int64
}

func (s *contextTestSuite) TestObject() {
	s.check = func(ctx *contextImpl) {
		v := &testStruct{}
		err := ctx.ArgObject(0, v)
		s.NoError(err)
		s.Equal(v, &testStruct{
			MyName: "name1",
			MyAge:  10,
		})
	}
	s.run(`test({"MyName":"name1", "MyAge": 10})`)
}

func (s *contextTestSuite) TestCallFunc() {
	s.eng.vm.Set("double", func(call otto.FunctionCall) otto.Value {
		f, _ := call.Argument(0).ToFloat()
		v, _ := otto.ToValue(f * 2)
		return v
	})
	s.check = func(ctx *contextImpl) {
		v, err := ctx.CallFunction(0, 3.3)
		s.NoError(err)
		s.Equal(v, 6.6)
	}
	s.run(`test(double)`)
}

func (s *contextTestSuite) TestLen() {
	s.check = func(ctx *contextImpl) {
		s.Equal(ctx.ArgLen(), 5)
	}
	s.run(`test("xxxx", 'yyy', 33, true, 'x')`)
}

func (s *contextTestSuite) TestReturn() {
	s.check = func(ctx *contextImpl) {
		ctx.Return(15)
		s.Equal(ctx.rtnValue, 15)
	}
	s.run(`test()`)
}

func (s *contextTestSuite) TestFuncReturn() {
	called := false
	s.check = func(ctx *contextImpl) {
		f := func(ctx sauron.PluginContext) error {
			called = true
			v, err := ctx.ArgFloat(0)
			s.NoError(err)
			s.Equal(v, 3.3)
			return nil
		}
		ctx.Return(sauron.FuncReturn(f))
	}
	s.run(`test()(3.3)`)
	s.True(called)
}

func (s *contextTestSuite) TestIsCallable() {
	s.check = func(ctx *contextImpl) {
		s.True(ctx.IsCallable(0))
		s.False(ctx.IsCallable(1))
	}
	s.run(`test(console.log, "xxx")`)
}

func (s *contextTestSuite) run(exp string) {
	s.eng.SetJobMeta(sauron.JobMeta{
		Script: exp,
	})
	s.NoError(s.eng.Run())
	s.True(s.called)
}
