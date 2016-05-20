package core

import (
	"encoding/json"
	"reflect"
	"sync"

	"straas.io/sauron"
)

// NewStore creates a store
func NewStore() (sauron.Store, error) {
	return &memStoreImpl{
		store: map[string][]byte{},
	}, nil
}

// memStoreImpl is a simple in-memory store implementation.
// It performs JSON encode before add to map, so i can change to other
// persistent storage (e.g. redis, file, SQL or NoSQL) quickly
// TODO: change to better encoding, JSON always stores number in float, so it's easy
// to lost precision for large integer
type memStoreImpl struct {
	// readwrite lock to allow concurrent access
	lock  sync.RWMutex
	store map[string][]byte
}

// Get returns data from store.
// Linke encoding/json, unmarshal parses the JSON-encoded data and stores the
// result in the value pointed to by v.
func (s *memStoreImpl) Get(ns, key string, v interface{}) (bool, error) {
	log.Debugf("[store] get ns:%s, key:%s", ns, key)
	data, ok := s.get(ns, key)
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal(data, v); err != nil {
		return false, err
	}
	return true, nil
}

// Set puts data into store
func (s *memStoreImpl) Set(ns, key string, v interface{}) error {
	log.Debugf("[store] set ns:%s, key:%s", ns, key)
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	k := s.makeKey(ns, key)

	// write lock
	s.lock.Lock()
	defer s.lock.Unlock()
	s.store[k] = data
	return nil
}

// Update updates data into store
func (s *memStoreImpl) Update(ns, key string, v interface{},
	action func(interface{}) (interface{}, error)) error {
	ok, err := s.Get(ns, key, v)
	if err != nil {
		return err
	}
	if ok {
		v, err = action(v)
	} else {
		v, err = action(nil)
	}
	if err != nil {
		return err
	}
	// nothing to do
	// golang nil comparion also includes types
	if v == nil || reflect.ValueOf(v).IsNil() {
		return nil
	}
	return s.Set(ns, key, v)
}

func (s *memStoreImpl) get(ns, key string) ([]byte, bool) {
	k := s.makeKey(ns, key)

	// read lock
	s.lock.RLock()
	defer s.lock.RUnlock()
	data, ok := s.store[k]
	return data, ok
}

// makeKey generates map key according to namespace and key
func (s *memStoreImpl) makeKey(ns, key string) string {
	// use JSON marshal
	bs, _ := json.Marshal([]string{ns, key})
	return string(bs)
}
