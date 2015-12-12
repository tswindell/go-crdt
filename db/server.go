package crdb

import (
    "net"
    "strconv"

    "google.golang.org/grpc"
    "golang.org/x/net/context"

    pb "github.com/tswindell/go-crdt/protos"
)

// The Server type is a concrete implementation of a CRDT network service.
type Server struct {
    listener *net.Listener
    service  *grpc.Server

    // Host/port information used to post service description.
    Hostname  string
    Port      int

    database *Database
}

// Returns a newly created Server instance.
func NewServer() *Server {
    d := new(Server)

    // Add credentials, and other options.
    d.service = grpc.NewServer()
    d.database = NewDatabase()

    // Register resource data types.
    tGSet := NewGSetResourceFactory(d.database)
    d.database.RegisterType(tGSet)

    t2PSet := NewTwoPhaseSetResourceFactory(d.database)
    d.database.RegisterType(t2PSet)

    // Register this instance as a CRDT service on our listener.
    pb.RegisterCRDTServer(d.service, d)

    pb.RegisterGrowOnlySetServer(d.service, tGSet)
    pb.RegisterTwoPhaseSetServer(d.service, t2PSet)
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

    LogInfo("Listening on %s:%d\n", hostname, port)

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

// The Create() server method
func (d *Server) Create(ctx context.Context, m *pb.CreateRequest) (*pb.CreateResponse, error) {
    status := &pb.Status{Success: true}

    resourceId, resourceKey, e := d.database.Create(ResourceType(m.ResourceType))
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    }

    LogInfo("CreateResponse: success=%v error=%s", status.Success, status.ErrorType)
    return &pb.CreateResponse{
               Status: status,
               ResourceId: string(resourceId),
               ResourceKey: string(resourceKey),
           }, nil
}

// The Attach() server method
func (d *Server) Attach(ctx context.Context, m *pb.AttachRequest) (*pb.AttachResponse, error) {
    resourceId := ResourceId(m.ResourceId)
    resourceKey := ResourceKey(m.ResourceKey)

    status := &pb.Status{Success: true}

    referenceId, e := d.database.Attach(resourceId, resourceKey)
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    }

    LogInfo("AttachResponse: success=%v error=%s", status.Success, status.ErrorType)
    return &pb.AttachResponse{
               Status: status,
               ReferenceId: string(referenceId),
           }, nil
}

// The Detach() server method
func (d *Server) Detach(ctx context.Context, m *pb.DetachRequest) (*pb.DetachResponse, error) {
    referenceId := ReferenceId(m.ReferenceId)
    status := &pb.Status{Success: true}

    e := d.database.Detach(referenceId)
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    }

    LogInfo("DetachResponse: success=%v error=%s", status.Success, status.ErrorType)
    return &pb.DetachResponse{Status: status}, nil
}

// The SupportedTypes() server method
func (d *Server) SupportedTypes(ctx context.Context, m *pb.EmptyMessage) (*pb.SupportedTypesResponse, error) {
    response := &pb.SupportedTypesResponse{Types: make([]*pb.TypeMessage, 0)}
    for _, v := range d.database.SupportedTypes() {
        response.Types = append(response.Types, &pb.TypeMessage{Type: string(v)})
    }
    return response, nil
}

// The IsSupportedType() server method
func (d *Server) IsSupportedType(ctx context.Context, m *pb.TypeMessage) (*pb.BooleanResponse, error) {
    return &pb.BooleanResponse{Value: d.database.IsSupportedType(ResourceType(m.Type))}, nil
}

