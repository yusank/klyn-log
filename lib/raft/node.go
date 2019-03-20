package raft

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

type raftNode struct {
	raft             *raft.Raft
	fsm              *FSM
	leaderNotifyChan chan bool
}

func newRaftTransport(cfg *config) (*raft.NetworkTransport, error) {
	addr, err := net.ResolveTCPAddr("tcp", cfg.raftTCPAddr)
	if err != nil {
		return nil, err
	}

	return raft.NewTCPTransport(addr.String(), addr, 3, 10*time.Second, os.Stderr)
}

func newRaftNode(cfg *config, ctx *raftContext) (*raftNode, error) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(cfg.raftTCPAddr)
	raftConfig.Logger = log.New(os.Stderr, "[raft]", log.LstdFlags)
	raftConfig.SnapshotInterval = 20 * time.Second
	raftConfig.SnapshotThreshold = 2
	leaderNotifyCh := make(chan bool, 1)
	raftConfig.NotifyCh = leaderNotifyCh

	trans, err := newRaftTransport(cfg)
	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(cfg.dataDir, 0700); err != nil {
		return nil, err
	}

	fsm := &FSM{
		ctx: ctx,
		log: log.New(os.Stderr, "[FSM]", log.LstdFlags),
	}

	snapshotStore, err := raft.NewFileSnapshotStore(cfg.dataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.dataDir, "raft-log.bolt"))
	if err != nil {
		return nil, err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.dataDir, "raft-stable.bolt"))
	if err != nil {
		return nil, err
	}

	newRaft, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, trans)
	if err != nil {
		return nil, err
	}

	if cfg.startAsLeader {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: trans.LocalAddr(),
				},
			},
		}

		newRaft.BootstrapCluster(configuration)
	}

	return &raftNode{raft: newRaft, fsm: fsm, leaderNotifyChan: leaderNotifyCh}, nil
}

func joinRaftCluster(cfg *config) error {
	url := fmt.Sprintf("http://%s/join?peerAddr=%s", cfg.joinAddr, cfg.raftTCPAddr)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Error joining cluster: %s", body))
	}

	return nil
}
