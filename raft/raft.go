// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package raft

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

type raftService struct {
	hService *httpServer
	cfg      *config
	cm       *cacheManager
	node     *raftNode
}

type raftContext struct {
	rs *raftService
}

func Server() {
	rs := &raftService{
		cfg: NewConfig(),
		cm:  NewCacheManager(),
	}

	ctx := &raftContext{rs: rs}

	var l net.Listener
	var err error
	l, err = net.Listen("tcp", rs.cfg.httpAddr)
	if err != nil {
		log.Fatal(fmt.Sprintf("listen %s failed: %s", rs.cfg.httpAddr, err))
	}
	log.Printf("http server listen:%s", l.Addr())

	logger := log.New(os.Stderr, "[httpServer]", log.Ldate|log.Ltime)
	httpServer := newHttpServer(ctx, logger)
	rs.hService = httpServer
	go func() {
		http.Serve(l, httpServer.mux)
	}()

	node, err := newRaftNode(rs.cfg, ctx)
	if err != nil {
		log.Fatal(fmt.Sprintf("new raft node failed:%v", err))
	}

	rs.node = node

	if rs.cfg.joinAddr != "" {
		err = joinRaftCluster(rs.cfg)
		if err != nil {
			log.Fatal(fmt.Sprintf("join raft cluster failed:%v", err))
		}
	}

	// monitor leadership
	for {
		select {
		case leader := <-rs.node.leaderNotifyChan:
			if leader {
				log.Println("become leader, enable write api")
				rs.hService.setWriteFlag(true)
			} else {
				log.Println("become follower, close write api")
				rs.hService.setWriteFlag(false)
			}
		}
	}

}
