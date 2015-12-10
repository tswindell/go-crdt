package crdtnet

import (
    "fmt"

    "google.golang.org/grpc"
    "golang.org/x/net/context"

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

func (d *Client) Close() {
    if d.connection == nil { return }
    d.connection.Close()
}

func (d *Client) CreateSet() (string, error) {
    r, e := d.CRDTClient.CreateSet(context.Background(), &pb.EmptyMessage{})
    if e != nil { return "", e }
    return r.SetId, nil
}

func (d *Client) DeleteSet(setId string) error {
    r, e := d.CRDTClient.DeleteSet(context.Background(),
                                   &pb.SetIdMessage{SetId: setId})
    if e != nil { return e }
    if !r.Success { return fmt.Errorf("Op Failed: %s", r.Mesg) }
    return nil
}

func (d *Client) ListSets() (chan string, error) {
    ch := make(chan string)
    r, e := d.CRDTClient.ListSets(context.Background(), &pb.EmptyMessage{})
    if e != nil { return nil, e }

    go func() {
        for {
            i, e := r.Recv()
            if e != nil { close(ch); break }
            ch<- i.SetId
        }
    }()

    return ch, nil
}

func (d *Client) GetObjects(setId string) (chan string, error) {
    ch := make(chan string)
    r, e := d.CRDTClient.GetObjects(context.Background(),
                                    &pb.SetIdMessage{SetId: setId})
    if e != nil { return ch, e }

    go func() {
        for {
            i, e := r.Recv()
            if e != nil { close(ch); break }
            ch<- i.Object
        }
    }()

    return ch, nil
}

func (d *Client) AddObject(setId, object string) (bool, error) {
    r, e := d.CRDTClient.AddObject(context.Background(),
                                   &pb.ObjectRequest{
                                       SetId: setId,
                                       Object: object,
                                   })
    if e != nil { return false, e }
    return r.Success, nil
}

func (d *Client) RemoveObject(setId, object string) (bool, error) {
    r, e := d.CRDTClient.RemoveObject(context.Background(),
                                      &pb.ObjectRequest{
                                          SetId: setId,
                                          Object: object,
                                      })
    if e != nil { return false, e }
    return r.Success, nil
}

func (d *Client) Contains(setId, object string) (bool, error) {
    r, e := d.CRDTClient.Contains(context.Background(),
                                  &pb.ObjectRequest{
                                      SetId: setId,
                                      Object: object,
                                  })
    if e != nil { return false, e }
    return r.Value, nil
}

func (d *Client) Equals(setA, setB string) (bool, error) {
    r, e := d.CRDTClient.Equals(context.Background(),
                                &pb.SetIdPairMessage{
                                    SetId1: setA,
                                    SetId2: setB,
                                })
    if e != nil { return false, e }
    return r.Value, nil
}

func (d *Client) Merge(setA, setB string) error {
    r, e := d.CRDTClient.Merge(context.Background(),
                              &pb.SetIdPairMessage{
                                  SetId1: setA,
                                  SetId2: setB,
                              })
    if e != nil { return e }
    if !r.Success { return fmt.Errorf("Op Failed: %s", r.Mesg) }
    return nil
}

func (d *Client) Clone(setA string) (string, error) {
    r, e := d.CRDTClient.Clone(context.Background(),
                               &pb.SetIdMessage{SetId: setA})
    if e != nil { return "", e }
    if len(r.SetId) == 0 { return "", fmt.Errorf("Op Failed") }
    return r.SetId, nil
}

