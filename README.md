# mini-s3

A stdlib-only, S3-compatible distributed object store in Go, built from scratch to actually understand how object storage internals work: content-addressed chunking, consistent hashing with virtual nodes, replication with write quorums, an append-only metadata log, mark-and-sweep garbage collection, and re-replication after node failures. It speaks enough of the S3 protocol (path-style addressing, XML responses, SigV4 auth, multipart upload) to be driven by `aws-cli`.

No third-party libraries — just `net/http`, `encoding/xml`, `crypto/*`, `sync`, and `os`.

## Architecture

Clients only ever talk to the coordinator. The coordinator owns all metadata and decides where chunks live; storage nodes are dumb, content-addressed blob stores:

```mermaid
flowchart LR
    Client(["aws-cli / S3 client"]) -->|"S3 API :9000"| Coordinator

    subgraph Coordinator["coordinator"]
        S3API["internal/s3api\nS3 routing + XML"] --> SigV4["internal/sigv4\nsignature check"]
        S3API --> Meta["internal/metadata\nbuckets, objects → chunk lists\nappend-only log"]
        S3API --> Chunker["internal/chunker\nsplit into chunks"]
        Chunker --> Repl["internal/replication\nN=3, W=2"]
        Repl --> Ring["internal/ring\nconsistent hash ring"]
        Ring --> Cluster["internal/cluster\nnode registry + health"]
    end

    Repl -->|"PUT/GET /chunks/{sha256}"| N1["storage node 1"]
    Repl --> N2["storage node 2"]
    Repl --> N3["storage node 3"]
    Repl --> N4["storage node 4"]
```

A `PutObject` writes every chunk to its replica set first and commits metadata last, so a crash can only ever leave orphan chunks (cleaned up by GC), never metadata pointing at missing data:

```mermaid
sequenceDiagram
    participant C as Client
    participant Co as Coordinator
    participant R as Ring
    participant N as Storage nodes (replica set)

    C->>Co: PUT /bucket/key (body stream)
    loop for each chunk read from the stream
        Co->>Co: sha256(chunk) = chunk ID
        Co->>R: Lookup(chunk ID, 3)
        R-->>Co: [node A, node B, node C]
        par write replicas
            Co->>N: PUT /chunks/{id}
        end
        N-->>Co: 2 of 3 acks (quorum)
    end
    Co->>Co: append object → chunk list to metadata log (fsync)
    Co-->>C: 200 OK, ETag
```
