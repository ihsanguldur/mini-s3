package s3api

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /{$}", h.ListBuckets)

	mux.HandleFunc("PUT /{bucket}", h.CreateBucket)
	mux.HandleFunc("PUT /{bucket}/{$}", h.CreateBucket)
	mux.HandleFunc("HEAD /{bucket}", h.HeadBucket)
	mux.HandleFunc("HEAD /{bucket}/{$}", h.HeadBucket)
	mux.HandleFunc("DELETE /{bucket}", h.DeleteBucket)
	mux.HandleFunc("DELETE /{bucket}/{$}", h.DeleteBucket)

	mux.HandleFunc("/", h.NotImplemented)
}
