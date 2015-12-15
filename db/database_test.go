package crdb

import "testing"
import "bytes"
import "crypto/rand"

func Test_NewResourceId_Valid(t *testing.T) {
    resourceId := NewResourceId("nil", "abcdef")
    if resourceId.GetStorageId() != "nil" {
        t.Errorf("Resource StorageId is wrong, %s != %s", "nil", resourceId.GetStorageId())
    }

    if resourceId.GetId() != "abcdef" {
        t.Errorf("Resource GetId is wrong, %s != %s", "abcdef", resourceId.GetId())
    }

    if !resourceId.IsValid() {
        t.Error("Resource is not valid!")
    }
}

func Test_NewResourceId_Invalid(t *testing.T) {
    resourceId := NewResourceId("", "")
    if resourceId.GetStorageId() != "" {
        t.Errorf("Resource StorageId is wrong, %s != %s", "", resourceId.GetStorageId())
    }

    if resourceId.GetId() != "" {
        t.Errorf("Resource GetId is wrong, %s != %s", "", resourceId.GetId())
    }

    if resourceId.IsValid() {
        t.Error("Resource is not meant to be valid!")
    }

    resourceId = NewResourceId("file", "")
    if resourceId.GetStorageId() != "file" {
        t.Errorf("Resource StorageId is wrong. (%s)", resourceId.GetStorageId())
    }

    if resourceId.GetId() != "" {
        t.Errorf("Resource Id is not empty: (%s)", resourceId.GetId())
    }

    if resourceId.IsValid() {
        t.Error("ResourceId is not meant to be valid!")
    }
}

func Test_NewReferenceId_Valid(t *testing.T) {
}

func Test_NewReferenceId_Invalid(t *testing.T) {
}

func Test_NewResourceKey_Valid(t *testing.T) {
    keydata := make([]byte, 16)
    _, _ = rand.Read(keydata)

    resourceKey := NewResourceKey("test-type", keydata)

    if resourceKey.TypeId() != "test-type" {
        t.Errorf("ResourceKey type is wrong, %s != %s", "test-type", resourceKey.TypeId())
    }

    if !bytes.Equal(resourceKey.KeyData(), keydata) {
        t.Error("ResourceKey data is wrong!")
    }

    if !resourceKey.IsValid() {
        t.Error("ResourceKey is invalid!")
    }
}

func Test_NewResourceKey_Invalid(t *testing.T) {
    resourceKey := NewResourceKey("", []byte{})

    if resourceKey.TypeId() != "" {
        t.Errorf("ResourceKey type is not empty! (%s)", resourceKey.TypeId())
    }

    if len(resourceKey.KeyData()) != 0 {
        t.Errorf("ResourceKey data is not empty! (%x)", resourceKey.KeyData())
    }

    if resourceKey.IsValid() {
        t.Error("ResourceKey is not meant to be valid!")
    }

    resourceKey = NewResourceKey("test", []byte{})
    if resourceKey.TypeId() != "test" {
        t.Errorf("ResourceKey type is wrong! (%s)", resourceKey.TypeId())
    }

    if !resourceKey.IsValid() {
        t.Error("ResourceKey is not valid!")
    }
}

func Test_NewDatabase_InitProcess(t *testing.T) {
    d := NewDatabase()

    if len(d.SupportedStorageTypes()) != 0 {
        t.Error("Empty support storage types check failed.")
    }

    if len(d.SupportedTypes()) != 0 {
        t.Error("Empty supported resource types check failed.")
    }

    if len(d.SupportedCryptoMethods()) != 0 {
        t.Error("Empty supported crypto methods check failed.")
    }

    if _, e := d.Create(ResourceType(""), "", ""); e != E_INVALID_TYPE {
        t.Error("Should not be able to make resources with no type!")
    }

    if _, e := d.Create(ResourceType("unknown"), "", ""); e != E_UNKNOWN_TYPE {
        t.Error("Should not be able to make resources with unknown type!")
    }

    fstore := NewFileStore("/tmp/crdb-store")
    if d.IsSupportedStorageType(fstore.TypeId()) {
        t.Error("IsSupportedStorageType check failed, should not be true!")
    }

    if e := d.RegisterStorageType(fstore); e != nil {
        t.Errorf("Failed to register storage type: %v", e)
    }

    if len(d.SupportedStorageTypes()) != 1 {
        t.Errorf("Expected 1 registed storage, got %d", len(d.SupportedStorageTypes()))
    }

    if !d.IsSupportedStorageType(fstore.TypeId()) {
        t.Error("IsSupportedStorageType check failed, should not be false!")
    }

    if e := d.RegisterStorageType(NewFileStore("")); e == nil {
        t.Error("We should not be able to add stores with same type.")
    }

    gset := NewGSetResourceFactory(d)
    if d.IsSupportedType(gset.TypeId()) {
        t.Error("IsSupportedResourceType check failed, should not be true!")
    }

    if e := d.RegisterType(gset); e != nil {
        t.Errorf("Failed to register resource type: %v", e)
    }

    if len(d.SupportedTypes()) != 1 {
        t.Errorf("Expected 1 registered resource type, got %d", len(d.SupportedTypes()))
    }

    if !d.IsSupportedType(gset.TypeId()) {
        t.Error("IsSupportedResourceType check failed, should not be false!")
    }

    if e := d.RegisterType(gset); e == nil {
        t.Error("We sbould not be able to add resource factories with same type!")
    }

    aes256cbc, _ := NewAESCryptoMethod(AES_256_KEY_SIZE)
    if d.IsSupportedCryptoMethod(aes256cbc.TypeId()) {
        t.Error("IsSupportedCryptoMethod check failed, shoult not be true!")
    }

    if e := d.RegisterCryptoMethod(aes256cbc); e != nil {
        t.Errorf("Failed to register crypto method: %v", e)
    }

    if len(d.SupportedCryptoMethods()) != 1 {
        t.Errorf("Expected 1 registed crypto method, got %d", len(d.SupportedCryptoMethods()))
    }

    if !d.IsSupportedCryptoMethod(aes256cbc.TypeId()) {
        t.Error("IsSupportedCryptoMethod check failed, should not be false.")
    }

    if e := d.RegisterCryptoMethod(aes256cbc); e == nil {
        t.Error("We should not be able to add resource factories with same type!")
    }
}

func Test_Database_Create(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewGSetResourceFactory(d)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorageType(NewFileStore("/tmp/crdb-test")); e != nil {
        t.Errorf("Failed to register storage: %v", e)
    }

    aes, e := NewAESCryptoMethod(AES_256_KEY_SIZE)
    if e != nil { t.Errorf("Failed to create crypto: %v", e) }

    if e := d.RegisterCryptoMethod(aes); e != nil {
        t.Errorf("Failed to register crypto: %v", e)
    }

    resource, e := d.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")
    if e != nil { t.Errorf("Failed to create new resource: %v", e) }
    if !resource.Id().IsValid() { t.Error("New resources' Id is not valid!") }
    if !resource.Key().IsValid() { t.Error("New resources' Key is not valid!") }
    if !resource.TypeId().IsValid() { t.Error("New resources' TypeId is not valid!") }

    if resource.TypeId() != ResourceType("crdt:gset") {
        t.Errorf("New resources' TypeId value incorrect: %s", resource.TypeId())
    }
}

func Test_Database_Create_Invalid(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewGSetResourceFactory(d)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorageType(NewFileStore("/tmp/crdb-test")); e != nil {
        t.Errorf("Failed to register storage: %v", e)
    }

    aes, e := NewAESCryptoMethod(AES_256_KEY_SIZE)
    if e != nil { t.Errorf("Failed to create crypto: %v", e) }

    if e := d.RegisterCryptoMethod(aes); e != nil {
        t.Errorf("Failed to register crypto: %v", e)
    }

    _, e = d.Create(ResourceType("invalid"), "file", "aes-256-cbc")
    if e == nil { t.Error("Invalid resource creation returned no error!") }
    if e != E_UNKNOWN_TYPE { t.Errorf("Incorrect error returned for invalid type: (%v)", e) }

    _, e = d.Create(ResourceType("crdt:gset"), "invalid", "aes-256-cbc")
    if e == nil { t.Error("Invalid resource creation returned no error!") }
    if e != E_INVALID_STORAGE { t.Errorf("Incorrect error returned for invalid storage: (%v)", e) }

    _, e = d.Create(ResourceType("crdt:gset"), "file", "invalid")
    if e == nil { t.Error("Invalid resource creation returned no error!") }
    if e != E_INVALID_CRYPTO { t.Errorf("Incorrect error returned for invalid crypto: (%v)", e ) }
}

func Test_Database_Attach(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewGSetResourceFactory(d)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorageType(NewFileStore("/tmp/crdb-test")); e != nil {
        t.Errorf("Failed to register storage: %v", e)
    }

    aes, e := NewAESCryptoMethod(AES_256_KEY_SIZE)
    if e != nil { t.Errorf("Failed to create crypto: %v", e) }

    if e := d.RegisterCryptoMethod(aes); e != nil {
        t.Errorf("Failed to register crypto: %v", e)
    }

    resource, e := d.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")
    if e != nil { t.Errorf("Failed to create resource: %v", e) }

    reference, e := d.Attach(resource.Id(), resource.Key())
    if e != nil { t.Errorf("Failed to attach resource: %v", e) }

    if !reference.IsValid() { t.Error("Got invalid reference!") }

    resourceId, e := d.Resolve(reference)
    if e != nil { t.Errorf("Failed to resolve reference: %v", e) }
    if resourceId != resource.Id() { t.Error("ResourceId check failed!") }
}

func Test_Database_Attach_Invalid(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewGSetResourceFactory(d)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorageType(NewFileStore("/tmp/crdb-test")); e != nil {
        t.Errorf("Failed to register storage: %v", e)
    }

    aes, e := NewAESCryptoMethod(AES_256_KEY_SIZE)
    if e != nil { t.Errorf("Failed to create crypto: %v", e) }

    if e := d.RegisterCryptoMethod(aes); e != nil {
        t.Errorf("Failed to register crypto: %v", e)
    }

    resource, e := d.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")
    if e != nil { t.Errorf("Failed to create resource: %v", e) }

    _, e = d.Attach(ResourceId(""), ResourceKey(""))
    if e == nil { t.Error("Invalid attach call returned no error!") }
    if e != E_INVALID_RESOURCE { t.Errorf("Incorrect error returned for invalid resource: %v", e) }

    _, e = d.Attach(resource.Id(), ResourceKey(""))
    if e == nil { t.Error("Invalid attach call returned no error!") }
    if e != E_INVALID_KEY { t.Errorf("Incorrect error returned for invalid key: %v", e) }
}

func Test_Database_Detach(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewGSetResourceFactory(d)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorageType(NewFileStore("/tmp/crdb-test")); e != nil {
        t.Errorf("Failed to register storage: %v", e)
    }

    aes, e := NewAESCryptoMethod(AES_256_KEY_SIZE)
    if e != nil { t.Errorf("Failed to create crypto: %v", e) }

    if e := d.RegisterCryptoMethod(aes); e != nil {
        t.Errorf("Failed to register crypto: %v", e)
    }

    resource, e := d.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")
    if e != nil { t.Errorf("Failed to create resource: %v", e) }

    reference, e := d.Attach(resource.Id(), resource.Key())
    if e != nil { t.Errorf("Failed to attach resource: %v", e) }

    e = d.Detach(reference)
    if e != nil { t.Error("Resource detach failed") }
}

func Test_Database_Detach_Invalid(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewGSetResourceFactory(d)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorageType(NewFileStore("/tmp/crdb-test")); e != nil {
        t.Errorf("Failed to register storage: %v", e)
    }

    aes, e := NewAESCryptoMethod(AES_256_KEY_SIZE)
    if e != nil { t.Errorf("Failed to create crypto: %v", e) }

    if e := d.RegisterCryptoMethod(aes); e != nil {
        t.Errorf("Failed to register crypto: %v", e)
    }

    resource, e := d.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")
    if e != nil { t.Errorf("Failed to create resource: %v", e) }

    _, e = d.Attach(resource.Id(), resource.Key())
    if e != nil { t.Errorf("Failed to attach resource: %v", e) }

    e = d.Detach(ReferenceId("invalid"))
    if e == nil { t.Error("Invalid detach returned no error!") }
    if e != E_INVALID_REFERENCE { t.Errorf("Expected invalid reference, got: %v", e) }
}

func Test_Database_Commit(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewGSetResourceFactory(d)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorageType(NewFileStore("/tmp/crdb-test")); e != nil {
        t.Errorf("Failed to register storage: %v", e)
    }

    aes, e := NewAESCryptoMethod(AES_256_KEY_SIZE)
    if e != nil { t.Errorf("Failed to create crypto: %v", e) }

    if e := d.RegisterCryptoMethod(aes); e != nil {
        t.Errorf("Failed to register crypto: %v", e)
    }

    resource, e := d.Create(ResourceType("crdt:gset"), "file", "aes-256-cbc")
    if e != nil { t.Errorf("Failed to create resource: %v", e) }

    reference, e := d.Attach(resource.Id(), resource.Key())
    if e != nil { t.Errorf("Failed to attach resource: %v", e) }

    e = d.Commit(reference)
    if e != nil { t.Errorf("Failed to commit resource: %v", e) }

    d = NewDatabase()
    if e := d.RegisterType(NewGSetResourceFactory(d)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorageType(NewFileStore("/tmp/crdb-test")); e != nil {
        t.Errorf("Failed to register storage: %v", e)
    }

    if e := d.RegisterCryptoMethod(aes); e != nil {
        t.Errorf("Failed to register crypto method: %v", e)
    }

    reference, e = d.Attach(resource.Id(), resource.Key())
    if e != nil { t.Errorf("Failed to attach resource: %v", e) }
}

