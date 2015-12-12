/*
 * Copyright (c) 2015 Tom Swindell (t.swindell@rubyx.co.uk)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package crdb

import (
    "fmt"
    "io"

    "google.golang.org/grpc"
    "golang.org/x/net/context"

    pb "github.com/tswindell/go-crdt/protos"
)

type GSetClient struct {
    pb.GrowOnlySetClient
}

func NewGSetClient(connection *grpc.ClientConn) *GSetClient {
    d := new(GSetClient)
    d.GrowOnlySetClient = pb.NewGrowOnlySetClient(connection)
    return d
}

// GSet API extensions to CRDB Client type
func (d *GSetClient) List(referenceId ReferenceId) (chan []byte, error) {
    r, e := d.GrowOnlySetClient.List(context.Background(),
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

func (d *GSetClient) Insert(referenceId ReferenceId, object []byte) error {
    r, e := d.GrowOnlySetClient.Insert(context.Background(),
                                       &pb.SetInsertRequest{
                                           Object: &pb.ResourceObject{
                                               ReferenceId: string(referenceId),
                                               Object: object,
                                           }})
    if e != nil { return e }
    if !r.Status.Success { return fmt.Errorf(r.Status.ErrorType) }
    return nil
}

func (d *GSetClient) Length(referenceId ReferenceId) (uint64, error) {
    r, e := d.GrowOnlySetClient.Length(context.Background(),
                                       &pb.SetLengthRequest{
                                           ReferenceId: string(referenceId),
                                       })
    if e != nil { return 0, e }
    if !r.Status.Success { return 0, fmt.Errorf(r.Status.ErrorType) }
    return r.Length, nil
}

func (d *GSetClient) Contains(referenceId ReferenceId, object []byte) (bool, error) {
    r, e := d.GrowOnlySetClient.Contains(context.Background(),
                                         &pb.SetContainsRequest{
                                             Object: &pb.ResourceObject{
                                                 ReferenceId: string(referenceId),
                                                 Object: object,
                                             }})
    if e != nil { return false, e }
    if !r.Status.Success { return false, fmt.Errorf(r.Status.ErrorType) }
    return r.Result, nil
}

func (d *GSetClient) Equals(referenceId, otherReferenceId ReferenceId) (bool, error) {
    r, e := d.GrowOnlySetClient.Equals(context.Background(),
                                       &pb.SetEqualsRequest{
                                           ReferenceId: string(referenceId),
                                           OtherReferenceId: string(otherReferenceId),
                                       })
    if e != nil { return false, e }
    if !r.Status.Success { return false, fmt.Errorf(r.Status.ErrorType) }
    return r.Result, nil
}

func (d *GSetClient) Merge(referenceId, otherReferenceId ReferenceId) error {
    r, e := d.GrowOnlySetClient.Merge(context.Background(),
                                       &pb.SetMergeRequest{
                                           ReferenceId: string(referenceId),
                                           OtherReferenceId: string(otherReferenceId),
                                       })
    if e != nil { return e }
    if !r.Status.Success { return fmt.Errorf(r.Status.ErrorType) }
    return nil
}

func (d *GSetClient) Clone(referenceId ReferenceId) (ResourceId, ResourceKey, error) {
    r, e := d.GrowOnlySetClient.Clone(context.Background(),
                                      &pb.SetCloneRequest{
                                          ReferenceId: string(referenceId),
                                      })
    if e != nil { return ResourceId(""), ResourceKey(""), e }
    if !r.Status.Success { return ResourceId(""), ResourceKey(""), fmt.Errorf(r.Status.ErrorType) }
    return ResourceId(r.ResourceId), ResourceKey(r.ResourceKey), nil
}

