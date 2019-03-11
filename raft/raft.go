// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package raft

type raftService struct {
	hService *httpServer
	cfg      *config
	cm       *cacheManager
	node     *raftNode
}

type raftContext struct {
	rs *raftService
}
