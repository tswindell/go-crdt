package crdb

import (
    "fmt"

    "google.golang.org/grpc"
    "golang.org/x/net/context"

    pb "github.com/tswindell/go-crdt/protos"
)

type Client struct {
    pb.CRDTClient
    GSetClient

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

    gsets := NewGSetClient(conn)
    d.GSetClient = *gsets

    return nil
}

func (d *Client) Close() {
    if d.connection == nil { return }
    d.connection.Close()
}

// The Create client request method
func (d *Client) Create(resourceType ResourceType) (ResourceId, ResourceKey, error) {
    r, e := d.CRDTClient.Create(context.Background(), &pb.CreateRequest{
                                                          ResourceType: string(resourceType),
                                                      })
    if e != nil { return ResourceId(""), ResourceKey(""), e }

    if !r.Status.Success {
        return ResourceId(""), ResourceKey(""), fmt.Errorf(r.Status.ErrorType)
    }

    return ResourceId(r.ResourceId), ResourceKey(r.ResourceKey), nil
}

// The Attach client request method
func (d *Client) Attach(resourceId ResourceId, resourceKey ResourceKey) (ReferenceId, error) {
    r, e := d.CRDTClient.Attach(context.Background(),
                                &pb.AttachRequest{
                                    ResourceId: string(resourceId),
                                    ResourceKey: string(resourceKey),
                                })
    if e != nil { return ReferenceId(""), e }

    if !r.Status.Success {
        return ReferenceId(""), fmt.Errorf(r.Status.ErrorType)
    }

    return ReferenceId(r.ReferenceId), nil
}

// The Detach client request method
func (d *Client) Detach(referenceId ReferenceId) error {
    r, e := d.CRDTClient.Detach(context.Background(),
                                &pb.DetachRequest{
                                    ReferenceId: string(referenceId),
                                })
    if e != nil { return e }
    if !r.Status.Success { return fmt.Errorf(r.Status.ErrorType) }
    return nil
}

// The SupportedTypes client request method
func (d *Client) SupportedTypes() ([]ResourceType, error) {
    results := make([]ResourceType, 0)

    r, e := d.CRDTClient.SupportedTypes(context.Background(), &pb.EmptyMessage{})
    if e != nil { return results, e }

    for _, v := range r.Types {
        results = append(results, ResourceType(v.Type))
    }

    return results, nil
}

// The IsSupportedType client request method
func (d *Client) IsSupportedType(resourceType ResourceType) (bool, error) {
    r, e := d.CRDTClient.IsSupportedType(context.Background(),
                                         &pb.TypeMessage{Type: string(resourceType)})
    if e != nil { return false, e }
    return r.Value, e
}

