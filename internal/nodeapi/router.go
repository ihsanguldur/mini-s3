package nodeapi

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("PUT /chunks/{id}", h.PutChunk)
	mux.HandleFunc("GET /chunks/{id}", h.GetChunk)
	mux.HandleFunc("HEAD /chunks/{id}", h.HeadChunk)
	mux.HandleFunc("DELETE /chunks/{id}", h.DeleteChunk)
}
