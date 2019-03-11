package raft

import (
	"encoding/json"
	"io"
	"sync"
)

type cacheManager struct {
	m sync.Map
}

func (cm *cacheManager) Get(key string) (val string, found bool) {
	v, ok := cm.m.Load(key)
	if !ok {
		return
	}

	val, found = v.(string)
	return
}

func (cm *cacheManager) Set(key, val string) {
	cm.m.Store(key, val)
}

func (cm *cacheManager) Marshal() ([]byte, error) {
	var m = make(map[string]interface{})

	cm.m.Range(func(key, value interface{}) bool {
		k := key.(string)
		m[k] = value
	})

	return json.Marshal(m)
}

func (cm *cacheManager) UnMarshal(serialized io.ReadCloser) error {
	var newData map[string]string
	if err := json.NewDecoder(serialized).Decode(&newData); err != nil {
		return err
	}

	cm.m = sync.Map{}
	for key, value := range newData {
		cm.m.Store(key, value)
	}

	return nil
}
