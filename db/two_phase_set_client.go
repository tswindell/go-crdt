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

