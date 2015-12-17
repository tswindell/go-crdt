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
    "reflect"

    "github.com/tswindell/go-crdt/sets"

    "google.golang.org/grpc"

    "golang.org/x/net/context"
    pb "github.com/tswindell/go-crdt/protos"
)

const (
    GROWONLYSET_RESOURCE_TYPE = ResourceType("crdt:gset")
    TWOPHASESET_RESOURCE_TYPE = ResourceType("crdt:2pset")
)

var (
    E_ALREADY_INSERTED = fmt.Errorf("crdt:error-already-inserted")
    E_ALREADY_REMOVED  = fmt.Errorf("crdt:error-already-removed")
)


// Generic ``Set'' function interfaces
type SetInsertInterface   interface { Insert(interface{}) bool     }
type SetRemoveInterface   interface { Remove(interface{}) bool     }
type SetContainsInterface interface { Contains(interface{}) bool   }
type SetLengthInterface   interface { Length() int                 }
type SetIterateInterface  interface { Iterate() <-chan interface{} }

type SerializeInterface   interface {

    Serialize(buff *bytes.Buffer) error

    Deserialize(buff *bytes.Buffer) error

}


// Generic ``Set'' resource type.
type SetResource struct {
    ResourceBase

    context interface{} // Polymorphic reference to concrete data type instance.
}

func (d *SetResource) Serialize(buff *bytes.Buffer) error {
    return d.context.(SerializeInterface).Serialize(buff)
}

func (d *SetResource) Deserialize(buff *bytes.Buffer) error {
    return d.context.(SerializeInterface).Deserialize(buff)
}


// The ResourceFactoryFunc type.
type ResourceFactoryFunc func(ResourceId, ResourceKey) Resource

// The NewGSetResource function adheres to ResourceFactoryFunc prototype.
func NewGSetResource(resourceId ResourceId, resourceKey ResourceKey) Resource {
    return &SetResource{
               ResourceBase{resourceId, resourceKey, GROWONLYSET_RESOURCE_TYPE},
               set.NewGSet(),
           }
}

// The New2PSetResource function adheres to ResourceFactoryFunc prototype.
func New2PSetResource(resourceId ResourceId, resourceKey ResourceKey) Resource {
    return &SetResource{
               ResourceBase{resourceId, resourceKey, TWOPHASESET_RESOURCE_TYPE},
               set.New2P(),
           }
}


// The SetResourceType type
type SetResourceType struct {
    database *Database
    typeId    ResourceType
    factory   ResourceFactoryFunc
}

func NewSetResourceType(database *Database,
                        typeId ResourceType,
                        factory ResourceFactoryFunc) *SetResourceType {

    return &SetResourceType{database, typeId, factory}
}

func (d *SetResourceType) TypeId() ResourceType { return d.typeId }

func (d *SetResourceType) Create(resourceId ResourceId, resourceKey ResourceKey) Resource {
    return d.factory(resourceId, resourceKey)
}

//TODO: FIXME?
func (d *SetResourceType) Equals(aResource, bResource Resource) (bool, error) {
    aSet := aResource.(*SetResource).context
    bSet := bResource.(*SetResource).context

    return reflect.ValueOf(aSet).MethodByName("Equals").Call([]reflect.Value{reflect.ValueOf(bSet)})[0].Bool(), nil
}

//TODO: FIXME?
func (d *SetResourceType) Merge(aResource, bResource Resource) error {
    aSet := aResource.(*SetResource).context
    bSet := bResource.(*SetResource).context

    reflect.ValueOf(aSet).MethodByName("Merge").Call([]reflect.Value{reflect.ValueOf(bSet)})

    return nil
}

//TODO: FIXME?
func (d *SetResourceType) Clone(resource Resource) (Resource, error) {
    newResource, e := d.database.Create(resource.Type(),
                                        resource.Key().TypeId(),
                                        resource.Id().GetStorageId())
    if e != nil { return nil, e }

    set := resource.(*SetResource).context
    newContext := reflect.ValueOf(set).MethodByName("Clone").Call([]reflect.Value{})[0].Interface()
    newResource.(*SetResource).context = newContext
    return newResource, nil
}

func (d *SetResourceType) Restore(resourceId ResourceId, resourceKey ResourceKey, buff *bytes.Buffer) (Resource, error) {
    resource := d.factory(resourceId, resourceKey)
    if e := resource.Deserialize(buff); e != nil { return nil, e }
    return resource, nil
}


// The SetResourceService type
type SetResourceService struct {
    database *Database
}

func NewSetResourceService(database *Database) *SetResourceService {
    return &SetResourceService{database}
}


// Concrete implementation for GSet List (It's either this, or we patch protoc output)
type GrowOnlySetService struct {SetResourceService}
func (d *GrowOnlySetService) List(m *pb.SetListRequest, stream pb.GrowOnlySet_ListServer) error {
    return d.SetResourceService.List(m, stream)
}


// Concrete implementation for 2PSet List ( ... )
type TwoPhaseSetService struct {SetResourceService}
func (d *TwoPhaseSetService) List(m *pb.SetListRequest, stream pb.TwoPhaseSet_ListServer) error {
    return d.SetResourceService.List(m, stream)
}


// The List() service method (abstract)
func (d *SetResourceService) List(m *pb.SetListRequest, stream grpc.ServerStream) error {
    r, e := d.database.Resolve(ReferenceId(m.ReferenceId))
    if e != nil { return nil }

    context := r.(*SetResource).context

    for v := range context.(SetIterateInterface).Iterate() {
        j, e := base64.StdEncoding.DecodeString(v.(string))
        if e != nil { return nil }

        stream.SendMsg(&pb.ResourceObject{
                           ReferenceId: string(m.ReferenceId),
                           Object: []byte(j),
                       })
    }

    return nil
}

// The Insert() service method
func (d *SetResourceService) Insert(ctx context.Context, m *pb.SetInsertRequest) (*pb.SetInsertResponse, error) {
    status := &pb.Status{Success: true}

    r, e := d.database.Resolve(ReferenceId(m.Object.ReferenceId))
    context := r.(*SetResource).context

    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()

    } else if !context.(SetInsertInterface).Insert(base64.StdEncoding.EncodeToString(m.Object.Object)) {
        status.Success = false
        status.ErrorType = E_ALREADY_INSERTED.Error()
    }

    return &pb.SetInsertResponse{Status: status}, nil
}

// The Remove() service method
func (d *SetResourceService) Remove(ctx context.Context, m *pb.SetRemoveRequest) (*pb.SetRemoveResponse, error) {
    r, e := d.database.Resolve(ReferenceId(m.Object.ReferenceId))
    if e != nil {
        return &pb.SetRemoveResponse{Status:&pb.Status{Success:false,ErrorType:e.Error()}}, nil
    }

    context := r.(*SetResource).context
    v := context.(SetRemoveInterface).Remove(base64.StdEncoding.EncodeToString(m.Object.Object))
    return &pb.SetRemoveResponse{Status:&pb.Status{Success:v}}, nil
}

// The Length() service method
func (d *SetResourceService) Length(ctx context.Context, m *pb.SetLengthRequest) (*pb.SetLengthResponse, error) {
    status := &pb.Status{Success: true}
    length := 0

    r, e := d.database.Resolve(ReferenceId(m.ReferenceId))
    context := r.(*SetResource).context

    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()

    } else {
        length = context.(SetLengthInterface).Length()
    }

    return &pb.SetLengthResponse{Status: status, Length: uint64(length)}, nil
}

// The Contains() service method
func (d *SetResourceService) Contains(ctx context.Context, m *pb.SetContainsRequest) (*pb.SetContainsResponse, error) {
    status := &pb.Status{Success: true}
    result := false

    r, e := d.database.Resolve(ReferenceId(m.Object.ReferenceId))
    context := r.(*SetResource).context

    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    } else {
        result = context.(SetContainsInterface).Contains(base64.StdEncoding.EncodeToString(m.Object.Object))
    }

    return &pb.SetContainsResponse{Status: status, Result: result}, nil
}

