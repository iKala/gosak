package encoding

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	yaml "gopkg.in/yaml.v2"
)

const (
	testJSONString = `
{
	"some": {
		"name": "aaa",
		"age": 10
	}
}
`
	testYAMLString = `
some:
  name: aaa
  age: 10
`
)

func TestNotification(t *testing.T) {
	suite.Run(t, new(notificationTestSuite))
}

type notificationTestSuite struct {
	suite.Suite
}

type testStruct struct {
	Some *RawMessage `json:"some" yaml:"some"`
}

type testSubSomeOne struct {
	Name string `json:"name" yaml:"name"`
}

type testSubSomeTwo struct {
	Age int `json:"age" yaml:"age"`
}

func (s *notificationTestSuite) TestJSON() {
	v := &testStruct{}
	one := &testSubSomeOne{}
	two := &testSubSomeTwo{}

	s.NoError(json.Unmarshal([]byte(testJSONString), v))
	s.NoError(v.Some.To(one))
	s.Equal(one.Name, "aaa")

	s.NoError(v.Some.To(two))
	s.Equal(two.Age, 10)
}

func (s *notificationTestSuite) TestYAML() {
	v := &testStruct{}
	one := &testSubSomeOne{}
	two := &testSubSomeTwo{}

	s.NoError(yaml.Unmarshal([]byte(testYAMLString), v))
	s.NoError(v.Some.To(one))
	s.Equal(one.Name, "aaa")

	s.NoError(v.Some.To(two))
	s.Equal(two.Age, 10)
}
