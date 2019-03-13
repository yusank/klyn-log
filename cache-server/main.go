package main

import (
	"sync"

	"git.yusank.cn/yusank/klyn-log/raft"
)

type cacheManager struct {
	m sync.Map
}

func main() {
	raft.Server()
}
