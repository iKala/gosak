package gosak

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONStringToMap(t *testing.T) {
	s := `{"hash":"ww9xPA7Abvwf8CTcih","name":"joe"}`
	r := JSONStringToMap(s)

	assert.Equal(t, 2, len(r))
	assert.Equal(t, "ww9xPA7Abvwf8CTcih", r["hash"])
}

func TestJSONStringToList(t *testing.T) {
	s := `[["10.240.0.80","10.240.0.12"],["10.240.0.113"]]`
	r := [][]string{}

	array := JSONStringToList(s)
	for _, subArr := range array {
		group := []string{}
		for _, ele := range subArr.([]interface{}) {
			group = append(group, ele.(string))
		}
		r = append(r, group)
	}

	assert.Equal(t, 2, len(r))
	assert.Equal(t, []string{"10.240.0.80", "10.240.0.12"}, r[0])
}

func TestJSONMapToString(t *testing.T) {
	m := map[string]interface{}{"name": "leo", "age": 10}

	r := JSONMapToString(m)

	assert.Equal(t, `{"age":10,"name":"leo"}`, string(r))
}

func TestPrettyPrintJSONMap(t *testing.T) {
	testedJSONMap := map[string]interface{}{"apple": 5, "lettuce": 7}

	PrettyPrintJSONMap(testedJSONMap)
}
