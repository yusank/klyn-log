package raft

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/hashicorp/raft"
)

const (
	EnableWriteNode uint32 = 0
	UnableWriteNode uint32 = 1
)

type httpServer struct {
	ctx         *raftContext
	mux         *http.ServeMux
	logger      *log.Logger
	enableWrite uint32
}

func newHttpServer(ctx *raftContext, l *log.Logger) *httpServer {
	mux := http.NewServeMux()
	hs := &httpServer{
		ctx:         ctx,
		mux:         mux,
		enableWrite: UnableWriteNode,
		logger:      l,
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
	h.logger.Printf("%s | %d", r.RequestURI, http.StatusOK)
	key := r.URL.Query().Get("key")
	val, ok := h.ctx.rs.cm.Get(key)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
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
	h.logger.Printf("%s | %d", r.RequestURI, http.StatusOK)
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

	event := logEntryData{Key: key, Value: val}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		h.logger.Printf("json.Marshal failed, err:%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "internal error\n")
		return
	}

	applyFuture := h.ctx.rs.node.raft.Apply(eventBytes, 5*time.Second)
	if err := applyFuture.Error(); err != nil {
		h.logger.Printf("raft.Apply failed:%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "internal error\n")
		return
	}

	w.Write([]byte("ok"))
	return
}

func (h *httpServer) joinHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Printf("%s | %d", r.RequestURI, http.StatusOK)
	peerAddr := r.URL.Query().Get("peerAddr")
	if peerAddr == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("invalid peerAddr"))
		return
	}

	if indexFeature := h.ctx.rs.node.raft.AddVoter(raft.ServerID(peerAddr), raft.ServerAddress(peerAddr), 0, 0); indexFeature.Error() != nil {
		h.logger.Println(indexFeature.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("add voter err"))
		return
	}

	w.Write([]byte("ok"))

}
