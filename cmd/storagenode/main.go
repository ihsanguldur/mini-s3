package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/ihsanguldur/mini-s3/internal/chunkstore"
	"github.com/ihsanguldur/mini-s3/internal/nodeapi"
)

func main() {
	addr := flag.String("addr", ":9101", "listening address")
	dataDir := flag.String("data-dir", "data/node1", "directory where chunks are stored")
	flag.Parse()

	store, err := chunkstore.New(*dataDir)
	if err != nil {
		log.Fatalf("open chunk store: %v", err)
	}

	mux := http.NewServeMux()
	nodeapi.RegisterRoutes(mux, nodeapi.NewHandler(store))

	log.Printf("storage node listening on %s, data dir %s", *addr, *dataDir)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
