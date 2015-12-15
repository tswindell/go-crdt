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
    "encoding/base64"

    "fmt"

    "github.com/tswindell/go-crdt/sets"

    "golang.org/x/net/context"
    pb "github.com/tswindell/go-crdt/protos"
)

const (
    TWOPHASESET_RESOURCE_TYPE = ResourceType("crdt:2pset")
)

var (
    E_ALREADY_REMOVED = fmt.Errorf("crdt:item-removed")
)

type TwoPhaseSetResource struct {
    resourceId ResourceId
    resourceKey ResourceKey

    object *set.TwoPhase
}

func NewTwoPhaseSetResource(resourceId ResourceId, resourceKey ResourceKey) *TwoPhaseSetResource {
    d := new(TwoPhaseSetResource)
    d.resourceId = resourceId
    d.resourceKey = resourceKey
    d.object = set.New2P()
    return d
}

func (d *TwoPhaseSetResource) Id() ResourceId {
    return d.resourceId
}

func (d *TwoPhaseSetResource) Key() ResourceKey {
    return d.resourceKey
}

func (d *TwoPhaseSetResource) TypeId() ResourceType {
    return TWOPHASESET_RESOURCE_TYPE
}

func (d *TwoPhaseSetResource) Serialize(buff *bytes.Buffer) error {
    return d.object.Serialize(buff)
}

type TwoPhaseSetResourceFactory struct {
    database  *Database
    resources  map[ResourceId]*TwoPhaseSetResource
}

func NewTwoPhaseSetResourceFactory(db *Database) *TwoPhaseSetResourceFactory {
    d := new(TwoPhaseSetResourceFactory)
    d.database  = db
    d.resources = make(map[ResourceId]*TwoPhaseSetResource)
    return d
}

func (d *TwoPhaseSetResourceFactory) TypeId() ResourceType {
    return TWOPHASESET_RESOURCE_TYPE
}

func (d *TwoPhaseSetResourceFactory) Create(resourceId ResourceId, resourceKey ResourceKey) Resource {
    resource := NewTwoPhaseSetResource(resourceId, resourceKey)
    d.resources[resourceId] = resource
    return resource
}

func (d *TwoPhaseSetResourceFactory) Restore(resourceId ResourceId, resourceKey ResourceKey, buff *bytes.Buffer) (Resource, error) {
    resource := NewTwoPhaseSetResource(resourceId, resourceKey)
    if e := resource.object.Deserialize(buff); e != nil { return nil, e }
    d.resources[resourceId] = resource
    return resource, nil
}

func (d *TwoPhaseSetResourceFactory) __resolve_reference(referenceId ReferenceId) (*TwoPhaseSetResource, error) {
    resourceId, e := d.database.Resolve(referenceId)
    if e != nil { return nil, e }

    resource, ok := d.resources[resourceId]
    if !ok { return nil, E_UNKNOWN_RESOURCE }

    return resource, nil
}

// The List() service method
func (d *TwoPhaseSetResourceFactory) List(m *pb.SetListRequest, stream pb.TwoPhaseSet_ListServer) error {
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
func (d *TwoPhaseSetResourceFactory) Insert(ctx context.Context, m *pb.SetInsertRequest) (*pb.SetInsertResponse, error) {
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

// The Remove() service method
func (d *TwoPhaseSetResourceFactory) Remove(ctx context.Context, m *pb.SetRemoveRequest) (*pb.SetRemoveResponse, error) {
    status := &pb.Status{Success: true}

    resource, e := d.__resolve_reference(ReferenceId(m.Object.ReferenceId))
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    } else if !resource.object.Remove(base64.StdEncoding.EncodeToString(m.Object.Object)) {
        status.Success = false
        status.ErrorType = E_ALREADY_REMOVED.Error()
    }

    return &pb.SetRemoveResponse{Status: status}, nil
}

// The Length() service method
func (d *TwoPhaseSetResourceFactory) Length(ctx context.Context, m *pb.SetLengthRequest) (*pb.SetLengthResponse, error) {
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
func (d *TwoPhaseSetResourceFactory) Contains(ctx context.Context, m *pb.SetContainsRequest) (*pb.SetContainsResponse, error) {
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
func (d *TwoPhaseSetResourceFactory) Equals(ctx context.Context, m *pb.SetEqualsRequest) (*pb.SetEqualsResponse, error) {
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
func (d *TwoPhaseSetResourceFactory) Merge(ctx context.Context, m *pb.SetMergeRequest) (*pb.SetMergeResponse, error) {
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
func (d *TwoPhaseSetResourceFactory) Clone(ctx context.Context, m *pb.SetCloneRequest) (*pb.SetCloneResponse, error) {
    resource, e := d.__resolve_reference(ReferenceId(m.ReferenceId))
    if e != nil {
        return &pb.SetCloneResponse{
                   Status: &pb.Status{Success: false, ErrorType: e.Error()},
               }, nil
    }

    cryptoId := resource.Key().TypeId()
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

