package raft

import "flag"

// 节点信息
type config struct {
	dataDir       string
	httpAddr      string
	raftTCPAddr   string
	startAsLeader bool
	joinAddr      string
}

func NewConfig() *config {
	var conf = &config{}
	var httpAddress = flag.String("http", "127.0.0.1:6000", "Http address")
	var raftTCPAddress = flag.String("raft", "127.0.0.1:7000", "raft tcp address")
	var node = flag.String("node", "node1", "raft node name")
	var bootstrap = flag.Bool("bootstrap", false, "start as raft cluster")
	var joinAddress = flag.String("join", "", "join address for raft cluster")
	flag.Parse()

	conf.dataDir = "./" + *node
	conf.httpAddr = *httpAddress
	conf.startAsLeader = *bootstrap
	conf.raftTCPAddr = *raftTCPAddress
	conf.joinAddr = *joinAddress

	return conf
}
