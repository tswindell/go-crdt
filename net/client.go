package crdtnet

import (
    "google.golang.org/grpc"

    pb "github.com/tswindell/go-crdt/protos"
)

type Client struct {
    pb.CRDTClient

    connection *grpc.ClientConn
}

func NewClient() *Client {
    d := new(Client)
    return d
}

func (d *Client) ConnectToHost(hostport string) error {
    conn, e := grpc.Dial(hostport, grpc.WithInsecure())
    if e != nil { return e }

    d.connection = conn
    d.CRDTClient = pb.NewCRDTClient(conn)

    return nil
}
