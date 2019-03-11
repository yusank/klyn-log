package main

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type cacheManager struct {
	m sync.Map
}

var cache = &cacheManager{m: sync.Map{}}

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

type CacheServer struct {
	cm *cacheManager
}

func (cs *CacheServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s | %s \n", r.Method, r.RequestURI)
	uri := strings.Split(r.RequestURI, "?")
	if uri[0] != "/cache" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		r.ParseForm()
		key := r.Form.Get("key")
		val, ok := cs.cm.Get(key)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, err := w.Write([]byte(val))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	case http.MethodPost:
		key := r.FormValue("key")
		val := r.FormValue("value")
		if key == "" || val == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		cs.cm.Set(key, val)
		_, err := w.Write([]byte("ok"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		w.WriteHeader(http.StatusNotFound)
	}

	return
}

func main() {
	s := &CacheServer{cm: cache}

	go func() {
		time.Sleep(3 * time.Second)

		http.Get("http://127.0.0.1:6000/cache?key=hello")
	}()

	log.Println("start listen")
	if err := http.ListenAndServe(":6000", s); err != nil {
		log.Fatal(err)
	}
}
