package main

import (
    "github.com/tswindell/go-crdt/net"
)

func main() {
    server := crdtnet.NewServer()
    server.Listen("127.0.0.1:9601")
}

