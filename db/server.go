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
    "net"
    "os/user"
    "path"
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
func NewServer() (*Server, error) {
    d := new(Server)

    // Add credentials, and other options.
    d.service = grpc.NewServer()
    d.database = NewDatabase()

    // Register persistent storage modules.
    u, e := user.Current()
    if e != nil { return nil, fmt.Errorf("Failed to get user") }
    filestore := NewFileStore(path.Join(u.HomeDir, ".crdb"))
    d.database.RegisterDatastore(filestore)

    // Register cryptographic methods.
    aes128cbc, _ := NewAESCryptoMethod(AES_128_KEY_SIZE)
    d.database.RegisterCryptoMethod(aes128cbc)

    aes192cbc, _ := NewAESCryptoMethod(AES_194_KEY_SIZE)
    d.database.RegisterCryptoMethod(aes192cbc)

    aes256cbc, _ := NewAESCryptoMethod(AES_256_KEY_SIZE)
    d.database.RegisterCryptoMethod(aes256cbc)

    // Register resource data types.
    tGSet := NewGSetResourceFactory(d.database)
    d.database.RegisterType(tGSet)

    t2PSet := NewTwoPhaseSetResourceFactory(d.database)
    d.database.RegisterType(t2PSet)

    // Register this instance as a CRDT service on our listener.
    pb.RegisterCRDTServer(d.service, d)

    pb.RegisterGrowOnlySetServer(d.service, tGSet)
    pb.RegisterTwoPhaseSetServer(d.service, t2PSet)
    return d, nil
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

    resourceId, resourceKey, e := d.database.Create(ResourceType(m.ResourceType),
                                                    m.StorageId,
                                                    m.CryptoId)
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

// The Commit() server method
func (d *Server) Commit(ctx context.Context, m *pb.CommitRequest) (*pb.CommitResponse, error) {
    referenceId := ReferenceId(m.ReferenceId)
    status := &pb.Status{Success: true}

    e := d.database.Commit(referenceId)
    if e != nil {
        status.Success = false
        status.ErrorType = e.Error()
    }

    LogInfo("CommitResponse: success=%v error=%s", status.Success, status.ErrorType)
    return &pb.CommitResponse{Status: status}, nil
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

// The SupportedStorageTypes() server method
func (d *Server) SupportedStorageTypes(ctx context.Context, m *pb.EmptyMessage) (*pb.SupportedStorageTypesResponse, error) {
    response := &pb.SupportedStorageTypesResponse{Types: make([]*pb.TypeMessage, 0)}
    for _, v := range d.database.SupportedStorageTypes() {
        response.Types = append(response.Types, &pb.TypeMessage{Type: v})
    }
    return response, nil
}

// The IsSupportedStorageType() server method
func (d *Server) IsSupportedStorageType(ctx context.Context, m *pb.TypeMessage) (*pb.BooleanResponse, error) {
    return &pb.BooleanResponse{Value: d.database.IsSupportedStorageType(m.Type)}, nil
}

// The SupportedCryptoMethods() server method
func (d *Server) SupportedCryptoMethods(ctx context.Context, m *pb.EmptyMessage) (*pb.SupportedCryptoMethodsResponse, error) {
    response := &pb.SupportedCryptoMethodsResponse{Types: make([]*pb.TypeMessage, 0)}
    for _, v := range d.database.SupportedCryptoMethods() {
        response.Types = append(response.Types, &pb.TypeMessage{Type: v})
    }
    return response, nil
}

// The IsSupportedCryptoMethod() server method
func (d *Server) IsSupportedCryptoMethod(ctx context.Context, m *pb.TypeMessage) (*pb.BooleanResponse, error) {
    return &pb.BooleanResponse{Value: d.database.IsSupportedCryptoMethod(m.Type)}, nil
}

