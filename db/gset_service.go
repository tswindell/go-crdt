package crdb

import (
    "encoding/base64"

    "fmt"

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

func (d *GSetResource) Type() ResourceType {
    return GSET_RESOURCE_TYPE
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

func (d *GSetResourceFactory) Type() ResourceType {
    return GSET_RESOURCE_TYPE
}

func (d *GSetResourceFactory) Create(resourceId ResourceId, resourceKey ResourceKey) Resource {
    resource := NewGSetResource(resourceId, resourceKey)
    d.resources[resourceId] = resource
    return resource
}

func (d *GSetResourceFactory) __resolve_reference(referenceId ReferenceId) (*set.GSet, error) {
    resourceId, e := d.database.Resolve(referenceId)
    if e != nil { return nil, e }

    resource, ok := d.resources[resourceId]
    if !ok { return nil, E_UNKNOWN_RESOURCE }

    return &resource.object, nil
}

// The List() service method
func (d *GSetResourceFactory) List(m *pb.SetListRequest, stream pb.GrowOnlySet_ListServer) error {
    gset, e := d.__resolve_reference(ReferenceId(m.ReferenceId))
    if e != nil { return nil }

    for v := range gset.Iterate() {
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

    gset, e := d.__resolve_reference(ReferenceId(m.Object.ReferenceId))
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    }

    if !gset.Insert(base64.StdEncoding.EncodeToString(m.Object.Object)) {
        status.Success = false
        status.ErrorType = E_ALREADY_PRESENT.Error()
    }

    return &pb.SetInsertResponse{Status: status}, nil
}

// The Length() service method
func (d *GSetResourceFactory) Length(ctx context.Context, m *pb.SetLengthRequest) (*pb.SetLengthResponse, error) {
    status := &pb.Status{Success: true}
    length := 0

    gset, e := d.__resolve_reference(ReferenceId(m.ReferenceId))
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()

    } else {
        length = gset.Length()
    }

    return &pb.SetLengthResponse{Status: status, Length: uint64(length)}, nil
}

// The Contains() service method
func (d *GSetResourceFactory) Contains(ctx context.Context, m *pb.SetContainsRequest) (*pb.SetContainsResponse, error) {
    status := &pb.Status{Success: true}
    result := false

    gset, e := d.__resolve_reference(ReferenceId(m.Object.ReferenceId))
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    } else {
        result = gset.Contains(base64.StdEncoding.EncodeToString(m.Object.Object))
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
               Result: setA.Equals(*setB),
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

    setA.Merge(*setB)

    return &pb.SetMergeResponse{
               Status: &pb.Status{Success: true},
           }, nil
}

// The Clone() service method
func (d *GSetResourceFactory) Clone(ctx context.Context, m *pb.SetCloneRequest) (*pb.SetCloneResponse, error) {
    set, e := d.__resolve_reference(ReferenceId(m.ReferenceId))
    if e != nil {
        return &pb.SetCloneResponse{
                   Status: &pb.Status{Success: false, ErrorType: e.Error()},
               }, nil
    }

    resourceId, resourceKey, e := d.database.Create(d.Type())
    if e != nil {
        return &pb.SetCloneResponse{
                   Status: &pb.Status{Success: false, ErrorType: e.Error()},
               }, nil
    }

    resource, _ := d.resources[resourceId]
    resource.object = set.Clone()

    return &pb.SetCloneResponse{
               Status: &pb.Status{Success: true},
               ResourceId: string(resourceId),
               ResourceKey: string(resourceKey),
           }, nil
}

