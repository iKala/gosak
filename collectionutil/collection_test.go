package collectionutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Person struct {
	name string
	age  int
}

func (p Person) IsQualified(...interface{}) bool {
	return p.age > 20
}

func TestGetQualifiedItems(t *testing.T) {
	people := []Person{
		{"apple", 10},
		{"ball", 20},
		{"cat", 50},
	}

	qualifiables := make([]Qualifiable, len(people))
	for i, v := range people {
		qualifiables[i] = v
	}

	count, qualifiedPeople := GetQualifiedItems(qualifiables)
	assert.Equal(t, 1, count)
	assert.Equal(t, "cat", qualifiedPeople[0].(Person).name)
}

func TestInStringSlice(t *testing.T) {
	slice := []string{"1.2.3.4", "5.6.7.8"}

	assert.False(t, InStringSlice(slice, "1.1.1.1"))
	assert.True(t, InStringSlice(slice, "5.6.7.8"))
}
