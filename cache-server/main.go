package main

import (
	"sync"

	"git.yusank.cn/yusank/klyn-log/lib/raft"
)

type cacheManager struct {
	m sync.Map
}

func main() {
	raft.Server()
}
