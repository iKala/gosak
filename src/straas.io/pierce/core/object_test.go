package core

import (
	"encoding/json"
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/suite"
)

func TestObject(t *testing.T) {
	suite.Run(t, new(objectTestSuite))
}

type objectTestSuite struct {
	suite.Suite
}

func (s *objectTestSuite) TestMarshal() {
	v := 123
	str, err := marshaller(v)
	s.NoError(err)
	s.Equal(str, "123")
}

func (s *objectTestSuite) TestUnmarshal() {
	str := `"123"`
	var v interface{}
	err := unmarshaller(str, &v)
	s.NoError(err)

	var exp interface{} = "123"
	s.Equal(v, exp)
}

func (s *objectTestSuite) TestSubkey() {
	skey, err := subkey("/aaa/bbb", "/aaa/bbb/ccc")
	s.NoError(err)
	s.Equal(skey, "ccc")

	skey, err = subkey("/aaa/bbb", "/aaa/bbb/ccc/ddd")
	s.NoError(err)
	s.Equal(skey, "ccc/ddd")

	skey, err = subkey("/aaa/bbb", "/aaa/bbb")
	s.NoError(err)
	s.Equal(skey, "")

	skey, err = subkey("/aaa/bbb", "/aaa/bbb/")
	s.NoError(err)
	s.Equal(skey, "")

	_, err = subkey("/xxx/bbb", "/aaa/bbb/ccc")
	s.Error(err)
}

func (s *objectTestSuite) TestToValue() {
	rawJSON := `{
		"key": "/aaa/bbb",
		"dir": true,
		"modifiedIndex": 123,
		"nodes": [
			{
				"key": "/aaa/bbb/ccc",
				"dir": false,
				"modifiedIndex": 123,
				"value": "111"
			},
			{
				"key": "/aaa/bbb/ddd",
				"dir": false,
				"modifiedIndex": 456,
				"value": "222"
			},
			{
				"key": "/aaa/bbb/eee",
				"dir": true,
				"modifiedIndex": 333,
				"nodes": [
					{
						"key": "/aaa/bbb/eee/fff",
						"dir": false,
						"modifiedIndex": 373,
						"value": "333"
					}				
				]
			}
		]
	}`

	var root *client.Node
	s.NoError(json.Unmarshal([]byte(rawJSON), &root))

	result, version, err := toValue(root, unmarshaller)
	s.NoError(err)
	s.Equal(version, uint64(456))
	s.Equal(result, map[string]interface{}{
		"ccc": float64(111),
		"ddd": float64(222),
		"eee": map[string]interface{}{
			"fff": float64(333),
		},
	})
}

func (s *objectTestSuite) TestToValueError() {
	rawJSON := `{
		"key": "/aaa/bbb",
		"dir": true,
		"modifiedIndex": 123,
		"nodes": [
			{
				"key": "/aaa/ddd/ccc",
				"dir": false,
				"modifiedIndex": 123,
				"value": "111"
			}
		]
	}`

	var root *client.Node
	s.NoError(json.Unmarshal([]byte(rawJSON), &root))

	_, _, err := toValue(root, unmarshaller)
	s.Error(err)
}

func (s *objectTestSuite) TestSetByPath() {
	var err error
	var v interface{} = map[string]interface{}{
		"ccc": 111,
		"ddd": 222,
		"eee": map[string]interface{}{
			"fff": 333,
		},
	}
	var exp interface{} = map[string]interface{}{
		"aaa": map[string]interface{}{
			"bbb": 10,
		},
		"ccc": 111,
		"ddd": 222,
		"eee": map[string]interface{}{
			"fff": 333,
			"ggg": 15,
		},
		"uuu": 33,
	}

	v, err = setByPath(v, "uuu", 33)
	s.NoError(err)
	v, err = setByPath(v, "aaa/bbb", 10)
	s.NoError(err)
	v, err = setByPath(v, "eee/ggg", 15)
	s.NoError(err)
	s.Equal(v, exp)

	// cannot convert int to map
	v, err = setByPath(v, "ddd/kkk", 15)
	s.Error(err)
	// v still ok
	s.Equal(v, exp)

	// replace root
	v, err = setByPath(v, "", 60)
	s.Equal(v, 60)
}

func (s *objectTestSuite) TestDelByPath() {
	var err error
	var v interface{} = map[string]interface{}{
		"aaa": map[string]interface{}{
			"bbb": 10,
		},
		"ccc": 111,
		"ddd": 222,
		"eee": map[string]interface{}{
			"fff": 333,
			"ggg": 15,
		},
		"uuu": 33,
	}
	var exp interface{} = map[string]interface{}{
		"aaa": map[string]interface{}{
			"bbb": 10,
		},
		"ccc": 111,
		"eee": map[string]interface{}{
			"ggg": 15,
		},
		"uuu": 33,
	}

	v, err = delByPath(v, "ddd")
	s.NoError(err)
	v, err = delByPath(v, "eee/fff")
	s.NoError(err)
	v, err = delByPath(v, "eee/kkk")
	s.NoError(err)
	v, err = delByPath(v, "bbb")
	s.NoError(err)
	s.Equal(v, exp)

	v, err = delByPath(v, "ccc/ddd")
	s.Error(err)
	// v still ok
	s.Equal(v, exp)

	v, err = delByPath(v, "")
	s.NoError(err)
	s.Nil(v)
}
