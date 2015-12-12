package crdb

import (
    "fmt"
    "io"

    "google.golang.org/grpc"
    "golang.org/x/net/context"

    pb "github.com/tswindell/go-crdt/protos"
)

type TwoPhaseSetClient struct {
    pb.TwoPhaseSetClient
}

func NewTwoPhaseSetClient(connection *grpc.ClientConn) *TwoPhaseSetClient {
    d := new(TwoPhaseSetClient)
    d.TwoPhaseSetClient = pb.NewTwoPhaseSetClient(connection)
    return d
}

// TwoPhase API extensions to CRDB Client type
func (d *TwoPhaseSetClient) List(referenceId ReferenceId) (chan []byte, error) {
    r, e := d.TwoPhaseSetClient.List(context.Background(),
                                     &pb.SetListRequest{
                                         ReferenceId: string(referenceId),
                                     })
    if e != nil { return nil, nil }

    ch := make(chan []byte) //TODO: Make buffered?
    go func() {
        for {
            object, e := r.Recv()
            if e == io.EOF { break }
            ch<- object.Object
        }
        close(ch)
    }()

    return ch, nil
}

func (d *TwoPhaseSetClient) Insert(referenceId ReferenceId, object []byte) error {
    r, e := d.TwoPhaseSetClient.Insert(context.Background(),
                                       &pb.SetInsertRequest{
                                           Object: &pb.ResourceObject{
                                               ReferenceId: string(referenceId),
                                               Object: object,
                                           }})
    if e != nil { return e }
    if !r.Status.Success { return fmt.Errorf(r.Status.ErrorType) }
    return nil
}

func (d *TwoPhaseSetClient) Remove(referenceId ReferenceId, object []byte) error {
    r, e := d.TwoPhaseSetClient.Remove(context.Background(),
                                       &pb.SetRemoveRequest{
                                           Object: &pb.ResourceObject{
                                               ReferenceId: string(referenceId),
                                               Object: object,
                                           }})
    if e != nil { return e }
    if !r.Status.Success { return fmt.Errorf(r.Status.ErrorType) }
    return nil
}

func (d *TwoPhaseSetClient) Length(referenceId ReferenceId) (uint64, error) {
    r, e := d.TwoPhaseSetClient.Length(context.Background(),
                                       &pb.SetLengthRequest{
                                           ReferenceId: string(referenceId),
                                       })
    if e != nil { return 0, e }
    if !r.Status.Success { return 0, fmt.Errorf(r.Status.ErrorType) }
    return r.Length, nil
}

func (d *TwoPhaseSetClient) Contains(referenceId ReferenceId, object []byte) (bool, error) {
    r, e := d.TwoPhaseSetClient.Contains(context.Background(),
                                         &pb.SetContainsRequest{
                                             Object: &pb.ResourceObject{
                                                 ReferenceId: string(referenceId),
                                                 Object: object,
                                             }})
    if e != nil { return false, e }
    if !r.Status.Success { return false, fmt.Errorf(r.Status.ErrorType) }
    return r.Result, nil
}

func (d *TwoPhaseSetClient) Equals(referenceId, otherReferenceId ReferenceId) (bool, error) {
    r, e := d.TwoPhaseSetClient.Equals(context.Background(),
                                       &pb.SetEqualsRequest{
                                           ReferenceId: string(referenceId),
                                           OtherReferenceId: string(otherReferenceId),
                                       })
    if e != nil { return false, e }
    if !r.Status.Success { return false, fmt.Errorf(r.Status.ErrorType) }
    return r.Result, nil
}

func (d *TwoPhaseSetClient) Merge(referenceId, otherReferenceId ReferenceId) error {
    r, e := d.TwoPhaseSetClient.Merge(context.Background(),
                                       &pb.SetMergeRequest{
                                           ReferenceId: string(referenceId),
                                           OtherReferenceId: string(otherReferenceId),
                                       })
    if e != nil { return e }
    if !r.Status.Success { return fmt.Errorf(r.Status.ErrorType) }
    return nil
}

func (d *TwoPhaseSetClient) Clone(referenceId ReferenceId) (ResourceId, ResourceKey, error) {
    r, e := d.TwoPhaseSetClient.Clone(context.Background(),
                                      &pb.SetCloneRequest{
                                          ReferenceId: string(referenceId),
                                      })
    if e != nil { return ResourceId(""), ResourceKey(""), e }
    if !r.Status.Success { return ResourceId(""), ResourceKey(""), fmt.Errorf(r.Status.ErrorType) }
    return ResourceId(r.ResourceId), ResourceKey(r.ResourceKey), nil
}

