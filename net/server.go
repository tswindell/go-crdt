package crdtnet

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"

    "fmt"
    "os"

    "net"
    "strconv"

    "google.golang.org/grpc"
    "golang.org/x/net/context"

    "github.com/tswindell/go-crdt/sets"

    pb "github.com/tswindell/go-crdt/protos"
)

// The Server type is a concrete implementation of a CRDT network service.
type Server struct {
    listener *net.Listener
    service  *grpc.Server

    Hostname  string
    Port      int

    database  map[string]*set.TwoPhase
}

// Returns a newly created Server instance.
func NewServer() *Server {
    d := new(Server)

    d.database = make(map[string]*set.TwoPhase)

    // TODO: Add credentials, and other options.
    d.service = grpc.NewServer()

    // Register this instance as a CRDT service on our listener.
    pb.RegisterCRDTServer(d.service, d)

    return d
}

// If successful starts server listening on hostport parameter.
func (d *Server) Listen(hostport string) error {
    listener, e := net.Listen("tcp", hostport)
    if e != nil { return e }
    d.listener = &listener

    hostname, port, e := __hostport_from_listener(d.listener)
    if e != nil { return e }
    d.Hostname = hostname
    d.Port = port

    // Start serving requests.
    return d.service.Serve(*d.listener)
}

func __hostport_from_listener(listener *net.Listener) (string, int, error) {
    host, p, e := net.SplitHostPort((*listener).Addr().String())
    if e != nil { return "", 0, e }

    port, e := strconv.Atoi(p)
    if e != nil { return "", 0, e }

    return host, port, nil
}

func __generate_random_key(size int) string {
    raw := make([]byte, size)
    rand.Read(raw)

    hash := sha256.New()
    hash.Write(raw[:])

    key := hex.EncodeToString(hash.Sum(nil))
    return key
}

// CRDT service implememtation -------------------------------------------------

func (d *Server) CreateSet(ctx context.Context, m *pb.EmptyMessage) (*pb.SetIdMessage, error) {
    key := __generate_random_key(16)

    s := set.New2P()
    d.database[key] = s

    fmt.Fprintf(os.Stderr, "Created new set with Id: %s\n", key)
    return &pb.SetIdMessage{SetId: key}, nil
}

func (d *Server) DeleteSet(ctx context.Context, m *pb.SetIdMessage) (*pb.StatusResponse, error) {
    _, ok := d.database[m.SetId]
    if !ok { return nil, fmt.Errorf("Failed to find set from Id: %s", m.SetId) }

    delete(d.database, m.SetId)

    fmt.Fprintf(os.Stderr, "Deleted set with Id: %s\n", m.SetId)
    return &pb.StatusResponse{Success: true}, nil
}

func (d *Server) ListSets(m *pb.EmptyMessage, s pb.CRDT_ListSetsServer) error {
    for k, _ := range d.database {
        if e := s.Send(&pb.SetIdMessage{SetId: k}); e != nil { return e }
    }
    return nil
}

func (d *Server) GetObjects(m *pb.SetIdMessage, s pb.CRDT_GetObjectsServer) error {
    set, ok := d.database[m.SetId]
    if !ok { return fmt.Errorf("Failed to find set from Id: %s", m.SetId) }

    items := set.ToSet()
    for i := range items.Iterate() {
        data, ok := i.(string)
        if !ok { continue }
        if e := s.Send(&pb.ObjectMessage{Object: data}); e != nil { return e }
    }
    return nil
}

func (d *Server) AddObject(ctx context.Context, m *pb.ObjectRequest) (*pb.StatusResponse, error) {
    set, ok := d.database[m.SetId]
    if !ok { return nil, fmt.Errorf("Failed to find set from Id: %s", m.SetId) }
    ok = set.Insert(m.Object)
    return &pb.StatusResponse{Success: ok}, nil
}

func (d *Server) RemoveObject(ctx context.Context, m *pb.ObjectRequest) (*pb.StatusResponse, error) {
    set, ok := d.database[m.SetId]
    if !ok { return nil, fmt.Errorf("Failed to find set from Id: %s", m.SetId) }
    ok = set.Remove(m.Object)
    return &pb.StatusResponse{Success: ok}, nil
}

func (d *Server) Contains(ctx context.Context, m *pb.ObjectRequest) (*pb.BooleanMessage, error) {
    set, ok := d.database[m.SetId]
    if !ok { return nil, fmt.Errorf("Failed to find set from Id: %s", m.SetId) }
    return &pb.BooleanMessage{Value: set.Contains(m.Object)}, nil
}

func (d *Server) Equals(ctx context.Context, m *pb.SetIdPairMessage) (*pb.BooleanMessage, error) {
    set1, ok := d.database[m.SetId1]
    if !ok { return nil, fmt.Errorf("Failed to find set from Id: %s", m.SetId1) }
    set2, ok := d.database[m.SetId2]
    if !ok { return nil, fmt.Errorf("Failed to find set from Id: %s", m.SetId2) }
    return &pb.BooleanMessage{Value: set1.Equals(set2)}, nil
}

func (d *Server) Merge(ctx context.Context, m *pb.SetIdPairMessage) (*pb.StatusResponse, error) {
    set1, ok := d.database[m.SetId1]
    if !ok { return nil, fmt.Errorf("Failed to find set from Id: %s", m.SetId1) }
    set2, ok := d.database[m.SetId2]
    if !ok { return nil, fmt.Errorf("Failed to find set from Id: %s", m.SetId2) }

    newSet := set1.Merge(set2)
    d.database[m.SetId1] = newSet

    return &pb.StatusResponse{Success: true}, nil
}

func (d *Server) Clone(ctx context.Context, m *pb.SetIdMessage) (*pb.SetIdMessage, error) {
    set, ok := d.database[m.SetId]
    if !ok { return nil, fmt.Errorf("Failed to find set from Id: %s", m.SetId) }

    newKey := __generate_random_key(16)

    newSet := set.Clone()
    d.database[newKey] = newSet

    fmt.Fprintf(os.Stderr, "Cloned set %s into new set with Id: %s\n", set, newKey)
    return &pb.SetIdMessage{SetId: newKey}, nil
}

