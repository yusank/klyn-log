package raft

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"
)

// FMS: finite state machine 有限状态机

type FSM struct {
	ctx *raftContext
}

type logEntryData struct {
	Key   string
	Value string
}

func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	e := logEntryData{}
	if err := json.Unmarshal(logEntry.Data, &e); err != nil {
		panic("Failed unmarshaling Raft log entry. This is a bug.")
	}
	f.ctx.rs.cm.Set(e.Key, e.Value)

	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{cm: f.ctx.rs.cm}, nil
}

func (f *FSM) Restore(serialized io.ReadCloser) error {
	return f.ctx.rs.cm.UnMarshal(serialized)
}
