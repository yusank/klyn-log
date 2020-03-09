package main

import (
	"sync"

	"github.com/yusank/klyn-log/lib/raft"
)

type cacheManager struct {
	m sync.Map
}

func main() {
	raft.Server()
}
