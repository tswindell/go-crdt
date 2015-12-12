package crdb

import (
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

func (d *TwoPhaseSetResource) Type() ResourceType {
    return TWOPHASESET_RESOURCE_TYPE
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

func (d *TwoPhaseSetResourceFactory) Type() ResourceType {
    return TWOPHASESET_RESOURCE_TYPE
}

func (d *TwoPhaseSetResourceFactory) Create(resourceId ResourceId, resourceKey ResourceKey) Resource {
    resource := NewTwoPhaseSetResource(resourceId, resourceKey)
    d.resources[resourceId] = resource
    return resource
}

func (d *TwoPhaseSetResourceFactory) __resolve_reference(referenceId ReferenceId) (*set.TwoPhase, error) {
    resourceId, e := d.database.Resolve(referenceId)
    if e != nil { return nil, e }

    resource, ok := d.resources[resourceId]
    if !ok { return nil, E_UNKNOWN_RESOURCE }

    return resource.object, nil
}

// The List() service method
func (d *TwoPhaseSetResourceFactory) List(m *pb.SetListRequest, stream pb.TwoPhaseSet_ListServer) error {
    set, e := d.__resolve_reference(ReferenceId(m.ReferenceId))
    if e != nil { return nil }

    for v := range set.Iterate() {
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

// The Remove() service method
func (d *TwoPhaseSetResourceFactory) Remove(ctx context.Context, m *pb.SetRemoveRequest) (*pb.SetRemoveResponse, error) {
    status := &pb.Status{Success: true}

    gset, e := d.__resolve_reference(ReferenceId(m.Object.ReferenceId))
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    }

    if !gset.Remove(base64.StdEncoding.EncodeToString(m.Object.Object)) {
        status.Success = false
        status.ErrorType = E_ALREADY_REMOVED.Error()
    }

    return &pb.SetRemoveResponse{Status: status}, nil
}

// The Length() service method
func (d *TwoPhaseSetResourceFactory) Length(ctx context.Context, m *pb.SetLengthRequest) (*pb.SetLengthResponse, error) {
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
func (d *TwoPhaseSetResourceFactory) Contains(ctx context.Context, m *pb.SetContainsRequest) (*pb.SetContainsResponse, error) {
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
               Result: setA.Equals(setB),
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

    setA.Merge(setB)

    return &pb.SetMergeResponse{
               Status: &pb.Status{Success: true},
           }, nil
}

// The Clone() service method
func (d *TwoPhaseSetResourceFactory) Clone(ctx context.Context, m *pb.SetCloneRequest) (*pb.SetCloneResponse, error) {
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

