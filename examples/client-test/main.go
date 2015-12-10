package main

import (
    "fmt"
    "os"

    "golang.org/x/net/context"

    "github.com/tswindell/go-crdt/net"
    pb "github.com/tswindell/go-crdt/protos"
)

func main() {
    client := crdtnet.NewClient()
    if e := client.ConnectToHost("127.0.0.1:9601"); e != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", e)
        os.Exit(1)
    }

    response, e := client.CreateSet(context.Background(), &pb.EmptyMessage{})
    if e != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", e)
        os.Exit(1)
    }

    setId := response.SetId
    fmt.Printf("New SetId: %s\n", setId)

    sets, e := client.ListSets(context.Background(), &pb.EmptyMessage{})
    if e != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", e)
        os.Exit(1)
    }

    for {
        i, e := sets.Recv()
        if e != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", e)
            break
        }

        fmt.Println("  Set: ", i.SetId)
    }

    for i := 0; i < 10; i++ {
        client.AddObject(context.Background(),
                         &pb.ObjectRequest{
                             SetId: setId,
                             Object: fmt.Sprintf("Object %d", i),
                         })
    }

    objs, e := client.GetObjects(context.Background(), &pb.SetIdMessage{SetId: setId})
    if e != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", e)
        os.Exit(1)
    }

    for {
        i, e := objs.Recv()
        if e != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", e)
            break
        }


        fmt.Println("  Object: ", i.Object)
    }

    os.Exit(0)
}

