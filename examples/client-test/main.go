package main

import (
    "fmt"
    "os"

    "github.com/tswindell/go-crdt/net"
)

func main() {
    client := crdtnet.NewClient()
    if e := client.ConnectToHost("127.0.0.1:9601"); e != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", e)
        os.Exit(1)
    }

    setId, e := client.CreateSet()
    if e != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", e)
        os.Exit(1)
    }
    fmt.Printf("New SetId: %s\n", setId)

    sets, e := client.ListSets()
    if e != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", e)
        os.Exit(1)
    }

    for setId := range sets {
        fmt.Println("  Set: ", setId)
    }

    for i := 0; i < 10; i++ {
        if _, e := client.AddObject(setId, fmt.Sprintf("Object %d", i)); e != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", e)
            break
        }
    }

    objs, e := client.GetObjects(setId)
    if e != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", e)
        os.Exit(1)
    }

    for obj := range objs {
        fmt.Println("  Object: ", obj)
    }

    os.Exit(0)
}

