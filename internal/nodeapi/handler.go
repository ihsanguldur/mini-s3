package nodeapi

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/ihsanguldur/mini-s3/internal/chunkstore"
)

type Handler struct {
	store *chunkstore.Store
}

func NewHandler(store *chunkstore.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) PutChunk(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Put(r.PathValue("id"), r.Body); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetChunk(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rc, size, err := h.store.Get(id)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	if _, err := io.Copy(w, rc); err != nil {
		log.Printf("get chunk %s: %v", id, err)
	}
}

func (h *Handler) HeadChunk(w http.ResponseWriter, r *http.Request) {
	size, err := h.store.Size(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteChunk(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
