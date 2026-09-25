package nodeapi

import (
	"errors"
	"log"
	"net/http"

	"github.com/ihsanguldur/mini-s3/internal/chunkstore"
)

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, chunkstore.ErrInvalidID), errors.Is(err, chunkstore.ErrHashMismatch):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, chunkstore.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		log.Printf("chunk store: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
