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

    "google.golang.org/grpc"
    "golang.org/x/net/context"

    pb "github.com/tswindell/go-crdt/protos"
)

type Client struct {
    pb.CRDTClient

    // Resource type declaration
    GSetClient
    TwoPhaseSetClient

    connection *grpc.ClientConn
}

// Returns new Client instance
func NewClient() *Client {
    d := new(Client)
    return d
}

// The ConnectToHost instance method
func (d *Client) ConnectToHost(hostport string) error {
    conn, e := grpc.Dial(hostport, grpc.WithInsecure())
    if e != nil { return e }

    d.connection = conn
    d.CRDTClient = pb.NewCRDTClient(conn)

    // Resource type registration.
    gsets := NewGSetClient(conn)
    d.GSetClient = *gsets

    tpsets := NewTwoPhaseSetClient(conn)
    d.TwoPhaseSetClient = *tpsets

    return nil
}

// The Close instance method
func (d *Client) Close() {
    if d.connection == nil { return }
    d.connection.Close()
}

// The Create client request method
func (d *Client) Create(resourceType ResourceType, storageId string, cryptoId string) (ResourceId, ResourceKey, error) {
    r, e := d.CRDTClient.Create(context.Background(), &pb.CreateRequest{
                                                          ResourceType: string(resourceType),
                                                          StorageId: storageId,
                                                          CryptoId: cryptoId,
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

// The Commit client request method
func (d *Client) Commit(referenceId ReferenceId) error {
    r, e := d.CRDTClient.Commit(context.Background(),
                                &pb.CommitRequest{
                                    ReferenceId: string(referenceId),
                                })
    if e != nil { return e }
    if !r.Status.Success { return fmt.Errorf(r.Status.ErrorType) }
    return nil
}

// The SupportedTypes client request method
func (d *Client) SupportedTypes() ([]string, error) {
    results := make([]string, 0)

    r, e := d.CRDTClient.SupportedTypes(context.Background(), &pb.EmptyMessage{})
    if e != nil { return results, e }

    for _, v := range r.Types {
        results = append(results, v.Type)
    }

    return results, nil
}

// The IsSupportedType client request method
func (d *Client) IsSupportedType(resourceType string) (bool, error) {
    r, e := d.CRDTClient.IsSupportedType(context.Background(),
                                         &pb.TypeMessage{Type: resourceType})
    if e != nil { return false, e }
    return r.Value, e
}

// The SupportedStorageTypes client request method
func (d *Client) SupportedStorageTypes() ([]string, error) {
    results := make([]string, 0)

    r, e := d.CRDTClient.SupportedStorageTypes(context.Background(), &pb.EmptyMessage{})
    if e != nil { return results, e }

    for _, v := range r.Types {
        results = append(results, v.Type)
    }

    return results, nil
}

// The IsSupportedStorageType client request method
func (d *Client) IsSupportedStorageType(storageType string) (bool, error) {
    r, e := d.CRDTClient.IsSupportedStorageType(context.Background(),
                                                &pb.TypeMessage{Type: storageType})
    if e != nil { return false, e }
    return r.Value, e
}

// The SupportedTypes client request method
func (d *Client) SupportedCryptoMethods() ([]string, error) {
    results := make([]string, 0)

    r, e := d.CRDTClient.SupportedCryptoMethods(context.Background(), &pb.EmptyMessage{})
    if e != nil { return results, e }

    for _, v := range r.Types {
        results = append(results, v.Type)
    }

    return results, nil
}

// The IsSupportedType client request method
func (d *Client) IsSupportedCryptoMethod(cryptoMethod string) (bool, error) {
    r, e := d.CRDTClient.IsSupportedCryptoMethod(context.Background(),
                                                 &pb.TypeMessage{Type: cryptoMethod})
    if e != nil { return false, e }
    return r.Value, e
}

