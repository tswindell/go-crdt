package crdb

import "testing"
import "fmt"
import "os"

var s *Server
var c *Client

func TestMain(m *testing.M) {
    var e error
    s, e = NewServer()
    if e != nil {
        fmt.Printf("Failed to instantiate new server: %v\n", e)
        os.Exit(1)
    }

    if e := s.Listen("127.0.0.1:0"); e != nil {
        fmt.Printf("Failed to bind service listener: %v\n", e)
        os.Exit(1)
    }

    c = NewClient()
    c.ConnectToHost(s.HostAddr())

    os.Exit(m.Run())
}

func Test_SupportedTypes(t *testing.T) {
    datatypes, e := c.SupportedTypes()
    if e != nil { t.Errorf("SupportedTypes failed: %v", e) }

    for _, dt := range datatypes {
        if ok, _ := c.IsSupportedType(dt); !ok { t.Error("IsSupportedType check failed!") }
    }

    ok, _ := c.IsSupportedType("invalid")
    if ok { t.Error("IsSupportedType check succeeded with invalid query") }
}

func Test_SupportedStorageTypes(t *testing.T) {
    stores, e := c.SupportedStorageTypes()
    if e != nil { t.Errorf("SupportedStorageTypes failed: %v", e) }

    for _, st := range stores {
        if ok, _ := c.IsSupportedStorageType(st); !ok {
            t.Error("IsSupportedStorageType check failed!")
        }
    }

    ok, _ := c.IsSupportedStorageType("invalid")
    if ok { t.Error("IsSupportedStorageType check succeeded with invalid query") }
}

func Test_SupportedCryptoMethods(t *testing.T) {
    cryptos, e := c.SupportedCryptoMethods()
    if e != nil { t.Errorf("SupportedCryptoMethods failed: %v") }

    for _, cm := range cryptos {
        if ok, _ := c.IsSupportedCryptoMethod(cm); !ok {
            t.Error("IsSupportedCryptoMethod check failed!")
        }
    }

    ok, _ := c.IsSupportedCryptoMethod("invalid")
    if ok { t.Error("IsSupportedCryptoMethod check succeeded with invalid query") }
}

func Test_Create(t *testing.T) {
    resourceId, resourceKey, e := c.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")
    if e != nil { t.Errorf("Create failed: %v", e) }

    if !resourceId.IsValid() { t.Error("Create returned invalid resource Id") }
    if !resourceKey.IsValid() { t.Error("Create returned invalid resource key") }
}

func Test_Create_Invalid(t *testing.T) {
    _, _, e := c.Create(ResourceType("invalid"), "file", "aes-256-cbc")
    if e == nil { t.Error("Create returned no error with invalid resource type.") }
    if e.Error() != E_UNKNOWN_TYPE.Error() { t.Errorf("Create returned wrong error: %v", e) }


    _, _, e = c.Create(ResourceType("crdt:gset"), "invalid", "aes-256-cbc")
    if e == nil { t.Error("Create returned no error with invalid storage id.") }
    if e.Error() != E_UNKNOWN_STORAGE.Error() { t.Errorf("Create returned wrong error: %v", e) }

    _, _, e = c.Create(ResourceType("crdt:gset"), "file", "invalid")
    if e == nil { t.Error("Create returned no error with invalid crypto.") }
    if e.Error() != E_UNKNOWN_CRYPTO.Error() { t.Errorf("Create returned wrong error: %v", e) }
}

func Test_Attach(t *testing.T) {
    resourceId, resourceKey, e := c.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")
    if e != nil { t.Errorf("Create returned error: %v", e) }

    referenceId, e := c.Attach(resourceId, resourceKey)
    if e != nil { t.Errorf("Attach returned error: %v", e) }
    if !referenceId.IsValid() { t.Error("Attach returned invalid reference Id") }
}

func Test_Attach_Invalid(t *testing.T) {
    _, e := c.Attach(ResourceId("invalid"), ResourceKey("invalid"))
    if e == nil { t.Error("Attach returned no error with invalid resource Id") }
    if e.Error() != E_UNKNOWN_RESOURCE.Error() { t.Errorf("Attach returned wrong error: %v", e) }

    resourceId, _, e := c.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")

    _, e = c.Attach(resourceId, ResourceKey("invalid"))
    if e == nil { t.Error("Attach returned no error with invalid resource key") }
    if e.Error() != E_INVALID_KEY.Error() { t.Errorf("Attach returned wrong error: %v", e) }
}

func Test_Detach(t *testing.T) {
    resourceId, resourceKey, e := c.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")
    if e != nil { t.Errorf("Failed to create resource: %v", e) }

    referenceId, e := c.Attach(resourceId, resourceKey)
    if e != nil { t.Errorf("Failed to attach resource: %v", e) }

    e = c.Detach(referenceId)
    if e != nil { t.Errorf("Failed to detach resource: %v", e) }
}

func Test_Detach_Invalid(t *testing.T) {
    e := c.Detach(ReferenceId("invalid"))
    if e == nil { t.Error("Detach returned no error with invalid reference id") }
    if e.Error() != E_INVALID_REFERENCE.Error() { t.Errorf("Detach returned incorrect error: %v", e) }
}

func Test_Commit(t *testing.T) {
    resourceId, resourceKey, e := c.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")
    if e != nil { t.Errorf("Failed to create resource: %v", e) }

    referenceId, e := c.Attach(resourceId, resourceKey)
    if e != nil { t.Errorf("Failed to attach resource: %v", e) }

    if e := c.Commit(referenceId); e != nil { t.Errorf("Commit failed: %v", e) }
}

func Test_Commit_Invalid(t *testing.T) {
    if e := c.Commit(ReferenceId("invalid")); e == nil { t.Error("Commit returned no error with invalid reference id") }
    if e := c.Commit(ReferenceId("invalid")); e.Error() != E_UNKNOWN_REFERENCE.Error() { t.Errorf("Commit returned wrong error: %v", e) }
}
