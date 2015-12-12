package main

import (
    "github.com/tswindell/go-crdt/db"
)

func main() {
    server := crdb.NewServer()
    server.Listen("127.0.0.1:9600")
}

