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
    "bytes"
    "fmt"

     "encoding/base64"

    "github.com/tswindell/go-crdt/sets"

    "golang.org/x/net/context"
    pb "github.com/tswindell/go-crdt/protos"
)

const (
    GSET_RESOURCE_TYPE = ResourceType("crdt:gset")
)

var (
    E_ALREADY_PRESENT = fmt.Errorf("crdt:item-already-present")
)

type GSetResource struct {
    resourceId ResourceId
    resourceKey ResourceKey

    object set.GSet
}

func NewGSetResource(resourceId ResourceId, resourceKey ResourceKey) *GSetResource {
    d := new(GSetResource)
    d.resourceId = resourceId
    d.resourceKey = resourceKey
    d.object = make(set.GSet)
    return d
}

func (d *GSetResource) Id() ResourceId {
    return d.resourceId
}

func (d *GSetResource) Key() ResourceKey {
    return d.resourceKey
}

func (d *GSetResource) TypeId() ResourceType {
    return GSET_RESOURCE_TYPE
}

func (d *GSetResource) Serialize(buff *bytes.Buffer) error {
    return d.object.Serialize(buff)
}

type GSetResourceFactory struct {
    database  *Database
    resources  map[ResourceId]*GSetResource
}

func NewGSetResourceFactory(db *Database) *GSetResourceFactory {
    d := new(GSetResourceFactory)
    d.database  = db
    d.resources = make(map[ResourceId]*GSetResource)
    return d
}

func (d *GSetResourceFactory) TypeId() ResourceType {
    return GSET_RESOURCE_TYPE
}

func (d *GSetResourceFactory) Create(resourceId ResourceId, resourceKey ResourceKey) Resource {
    resource := NewGSetResource(resourceId, resourceKey)
    d.resources[resourceId] = resource
    return resource
}

func (d *GSetResourceFactory) Restore(resourceId ResourceId, resourceKey ResourceKey, buff *bytes.Buffer) (Resource, error) {
    resource := NewGSetResource(resourceId, resourceKey)
    if e := resource.object.Deserialize(buff); e != nil { return nil, e }
    d.resources[resourceId] = resource
    return resource, nil
}

func (d *GSetResourceFactory) __resolve_reference(referenceId ReferenceId) (*GSetResource, error) {
    resourceId, e := d.database.Resolve(referenceId)
    if e != nil { return nil, e }

    resource, ok := d.resources[resourceId]
    if !ok { return nil, E_UNKNOWN_RESOURCE }

    return resource, nil
}

// The List() service method
func (d *GSetResourceFactory) List(m *pb.SetListRequest, stream pb.GrowOnlySet_ListServer) error {
    resource, e := d.__resolve_reference(ReferenceId(m.ReferenceId))
    if e != nil { return nil }

    for v := range resource.object.Iterate() {
        j, e := base64.StdEncoding.DecodeString(v.(string))
        if e != nil { return nil }

        stream.Send(&pb.ResourceObject{
                        ReferenceId: string(m.ReferenceId),
                        Object: []byte(j),
                    })
    }

    return nil
}

// The Insert() service method
func (d *GSetResourceFactory) Insert(ctx context.Context, m *pb.SetInsertRequest) (*pb.SetInsertResponse, error) {
    status := &pb.Status{Success: true}

    resource, e := d.__resolve_reference(ReferenceId(m.Object.ReferenceId))
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    } else if !resource.object.Insert(base64.StdEncoding.EncodeToString(m.Object.Object)) {
        status.Success = false
        status.ErrorType = E_ALREADY_PRESENT.Error()
    }

    return &pb.SetInsertResponse{Status: status}, nil
}

// The Length() service method
func (d *GSetResourceFactory) Length(ctx context.Context, m *pb.SetLengthRequest) (*pb.SetLengthResponse, error) {
    status := &pb.Status{Success: true}
    length := 0

    resource, e := d.__resolve_reference(ReferenceId(m.ReferenceId))
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()

    } else {
        length = resource.object.Length()
    }

    return &pb.SetLengthResponse{Status: status, Length: uint64(length)}, nil
}

// The Contains() service method
func (d *GSetResourceFactory) Contains(ctx context.Context, m *pb.SetContainsRequest) (*pb.SetContainsResponse, error) {
    status := &pb.Status{Success: true}
    result := false

    resource, e := d.__resolve_reference(ReferenceId(m.Object.ReferenceId))
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    } else {
        result = resource.object.Contains(base64.StdEncoding.EncodeToString(m.Object.Object))
    }

    return &pb.SetContainsResponse{Status: status, Result: result}, nil
}

// The Equals() service method
func (d *GSetResourceFactory) Equals(ctx context.Context, m *pb.SetEqualsRequest) (*pb.SetEqualsResponse, error) {
    setA, e := d.__resolve_reference(ReferenceId(m.ReferenceId))
    if e != nil {
        return &pb.SetEqualsResponse{
                   Status: &pb.Status{Success: false, ErrorType: e.Error()},
               }, nil
    }

    setB, e := d.__resolve_reference(ReferenceId(m.OtherReferenceId))
    if e != nil {
        return &pb.SetEqualsResponse{
                   Status: &pb.Status{Success: false, ErrorType: e.Error()},
               }, nil
    }

    return &pb.SetEqualsResponse{
               Status: &pb.Status{Success: true},
               Result: setA.object.Equals(setB.object),
           }, nil
}

// The Merge() service method
func (d *GSetResourceFactory) Merge(ctx context.Context, m *pb.SetMergeRequest) (*pb.SetMergeResponse, error) {
    setA, e := d.__resolve_reference(ReferenceId(m.ReferenceId))
    if e != nil {
        return &pb.SetMergeResponse{
                   Status: &pb.Status{Success: false, ErrorType: e.Error()},
               }, nil
    }

    setB, e := d.__resolve_reference(ReferenceId(m.OtherReferenceId))
    if e != nil {
        return &pb.SetMergeResponse{
                   Status: &pb.Status{Success: false, ErrorType: e.Error()},
               }, nil
    }

    setA.object.Merge(setB.object)

    return &pb.SetMergeResponse{
               Status: &pb.Status{Success: true},
           }, nil
}

// The Clone() service method
func (d *GSetResourceFactory) Clone(ctx context.Context, m *pb.SetCloneRequest) (*pb.SetCloneResponse, error) {
    resource, e := d.__resolve_reference(ReferenceId(m.ReferenceId))
    if e != nil {
        return &pb.SetCloneResponse{
                   Status: &pb.Status{Success: false, ErrorType: e.Error()},
               }, nil
    }

    cryptoId  := resource.Key().TypeId()
    storageId := resource.Id().GetStorageId()

    newResourceId, newResourceKey, e := d.database.Create(d.TypeId(), cryptoId, storageId)
    if e != nil {
        return &pb.SetCloneResponse{
                   Status: &pb.Status{Success: false, ErrorType: e.Error()},
               }, nil
    }

    newResource, _ := d.resources[newResourceId]
    newResource.object = resource.object.Clone()

    return &pb.SetCloneResponse{
               Status: &pb.Status{Success: true},
               ResourceId: string(newResourceId),
               ResourceKey: string(newResourceKey),
           }, nil
}

