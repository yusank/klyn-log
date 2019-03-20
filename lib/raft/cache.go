package raft

import (
	"encoding/json"
	"io"
	"log"
	"sync"
)

type cacheManager struct {
	m sync.Map
}

func NewCacheManager() *cacheManager {
	return &cacheManager{m: sync.Map{}}
}

func (cm *cacheManager) Get(key string) (val string, found bool) {
	v, ok := cm.m.Load(key)
	log.Println(key, v, ok)
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
		log.Println("marshall", key, value)
		m[k] = value

		return true
	})

	return json.Marshal(m)
}

func (cm *cacheManager) UnMarshal(serialized io.ReadCloser) error {
	log.Println("unmarshall")
	var newData map[string]string
	if err := json.NewDecoder(serialized).Decode(&newData); err != nil {
		return err
	}

	cm.m = sync.Map{}
	for key, value := range newData {
		log.Println(key, value)
		cm.m.Store(key, value)
	}

	return nil
}
