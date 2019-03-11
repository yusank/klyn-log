package raft

import (
	"net/http"
	"sync/atomic"

	"github.com/hashicorp/raft"
)

const (
	EnableWriteNode uint32 = 0
	UnableWriteNode uint32 = 1
)

type httpServer struct {
	ctx         *raftContext
	mux         *http.ServeMux
	enableWrite uint32
}

func newHttpServer(ctx *raftContext) *httpServer {
	mux := http.NewServeMux()
	hs := &httpServer{
		ctx:         ctx,
		mux:         mux,
		enableWrite: UnableWriteNode,
	}

	mux.HandleFunc("/get", hs.getHandler)
	mux.HandleFunc("/set", hs.setHandler)
	mux.HandleFunc("/join", hs.joinHandler)

	return hs
}

func (h *httpServer) enableWriteNode() bool {
	return atomic.LoadUint32(&h.enableWrite) == EnableWriteNode
}

func (h *httpServer) setWriteFlag(enable bool) {
	if enable {
		atomic.StoreUint32(&h.enableWrite, EnableWriteNode)
	} else {
		atomic.StoreUint32(&h.enableWrite, UnableWriteNode)
	}
}

func (h *httpServer) getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.Form.Get("key")
	val, ok := h.ctx.rs.cm.Get(key)

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
}

func (h *httpServer) setHandler(w http.ResponseWriter, r *http.Request) {
	if !h.enableWriteNode() {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	key := r.FormValue("key")
	val := r.FormValue("value")
	if key == "" || val == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.ctx.rs.cm.Set(key, val)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

func (h *httpServer) joinHandler(w http.ResponseWriter, r *http.Request) {
	peerAddr := r.URL.Query().Get("peerAddr")
	if peerAddr == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("invalid peerAddr"))
		return
	}

	if err := h.ctx.rs.node.raft.AddVoter(raft.ServerID(peerAddr), raft.ServerAddress(peerAddr), 0, 0); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("add voter err"))
		return
	}

	w.Write([]byte("ok"))
}
