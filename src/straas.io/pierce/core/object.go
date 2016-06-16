package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/etcd/client"
)

// extractKey remoevs prefix from etcd key
func extractKey(prefix, etcdKey string) (string, error) {
	if !strings.HasPrefix(etcdKey, prefix) {
		return "", fmt.Errorf("unexcepted etcd key %s for %s", etcdKey, prefix)
	}
	return strings.TrimPrefix(etcdKey[len(prefix):], "/"), nil
}

// unmarshaller
func unmarshaller(value string, v interface{}) error {
	return json.Unmarshal([]byte(value), v)
}

// marshaller
func marshaller(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// setByPath is like loadsh _.set, it sets the value at path of object.
// If a portion of path doesn’t exist, it’s created
// For example, if object is {"a": {"b": 10}}
// obj, _ = set(obj, "a/c/d", 5) => {"a": {"b": 10, "c":{"d": 5}}}
func setByPath(root interface{}, path string, v interface{}) (interface{}, error) {
	if path == "" {
		return v, nil
	}

	var nested map[string]interface{}
	root, nested, err := ensureMap(root, true)
	if err != nil {
		return nil, err
	}
	keys := strings.Split(path, "/")
	lastIdx := len(keys) - 1

	for i, key := range keys {
		if i == lastIdx {
			nested[key] = v
			break
		}
		var cv map[string]interface{}
		nested[key], cv, err = ensureMap(nested[key], true)
		if err != nil {
			return nil, err
		}
		nested = cv
	}
	return root, nil
}

// delByPath deletes the path of object.
// For example, if object is {"a": {"b": 10}}
// obj, _ = del(obj, "a/b") => {"a": {}}
func delByPath(root interface{}, path string) (interface{}, error) {
	if path == "" {
		return nil, nil
	}
	var nested map[string]interface{}
	_, nested, err := ensureMap(root, false)
	if err != nil {
		return nil, err
	}
	if nested == nil {
		return root, nil
	}

	keys := strings.Split(path, "/")
	lastIdx := len(keys) - 1

	for i, key := range keys {
		if i == lastIdx {
			delete(nested, key)
			break
		}
		var cv map[string]interface{}
		_, cv, err = ensureMap(nested[key], false)
		if err != nil {
			return nil, err
		}
		if cv == nil {
			break
		}
		nested = cv
	}
	return root, nil
}

// ensureMap converts container to map if possible or creates map if necessary
func ensureMap(container interface{},
	create bool) (interface{}, map[string]interface{}, error) {
	var nested map[string]interface{}
	var ok bool

	if container == nil {
		if !create {
			return container, nil, nil
		}
		nested = map[string]interface{}{}
		container = nested
	} else if nested, ok = container.(map[string]interface{}); !ok {
		return nil, nil, fmt.Errorf("cannot convert container to map")
	}
	return container, nested, nil
}

func toValue(node *client.Node,
	unmarshaller func(string, interface{}) error) (interface{}, uint64, error) {
	if !node.Dir {
		if node.Value == "" {
			return nil, node.ModifiedIndex, nil
		}
		var v interface{}
		if err := unmarshaller(node.Value, &v); err != nil {
			return nil, 0, err
		}
		return v, node.ModifiedIndex, nil
	}

	vs := map[string]interface{}{}
	maxIndex := node.ModifiedIndex

	for _, n := range node.Nodes {
		key, err := subkey(node.Key, n.Key)
		if err != nil {
			return nil, 0, err
		}
		v, idx, err := toValue(n, unmarshaller)
		if err != nil {
			return nil, 0, err
		}
		if idx > maxIndex {
			maxIndex = idx
		}
		vs[key] = v
	}
	return vs, maxIndex, nil
}

func subkey(prefix, etcdKey string) (string, error) {
	if !strings.HasPrefix(etcdKey, prefix) {
		return "", fmt.Errorf("unexcepted etcd key %s for %s", etcdKey, prefix)
	}
	return strings.TrimPrefix(etcdKey[len(prefix):], "/"), nil
}
