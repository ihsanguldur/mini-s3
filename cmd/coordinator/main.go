package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ihsanguldur/mini-s3/internal/metadata"
	"github.com/ihsanguldur/mini-s3/internal/s3api"
)

func main() {
	addr := flag.String("addr", ":9000", "listening address")
	flag.Parse()

	meta := metadata.New()
	mux := http.NewServeMux()
	s3api.RegisterRoutes(mux, s3api.NewHandler(meta))

	log.Printf("coordinator listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
