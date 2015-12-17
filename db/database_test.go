package crdb

import "testing"

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

    fstore := NewFileStore("/tmp/crdb-test")
    if d.IsSupportedStorageType(fstore.TypeId()) {
        t.Error("IsSupportedStorageType check failed, should not be true!")
    }

    if e := d.RegisterStorage(fstore); e != nil {
        t.Errorf("Failed to register storage type: %v", e)
    }

    if len(d.SupportedStorageTypes()) != 1 {
        t.Errorf("Expected 1 registed storage, got %d", len(d.SupportedStorageTypes()))
    }

    if !d.IsSupportedStorageType(fstore.TypeId()) {
        t.Error("IsSupportedStorageType check failed, should not be false!")
    }

    if e := d.RegisterStorage(NewFileStore("")); e == nil {
        t.Error("We should not be able to add stores with same type.")
    }

    gset := NewSetResourceType(d, GROWONLYSET_RESOURCE_TYPE, NewGSetResource)
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

    if e := d.RegisterType(NewSetResourceType(d, GROWONLYSET_RESOURCE_TYPE, NewGSetResource)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorage(NewFileStore("/tmp/crdb-test")); e != nil {
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
    if !resource.Type().IsValid() { t.Error("New resources' TypeId is not valid!") }

    if resource.Type() != ResourceType("crdt:gset") {
        t.Errorf("New resources' TypeId value incorrect: %s", resource.Type())
    }
}

func Test_Database_Create_Invalid(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewSetResourceType(d, GROWONLYSET_RESOURCE_TYPE, NewGSetResource)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorage(NewFileStore("/tmp/crdb-test")); e != nil {
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
    if e != E_UNKNOWN_STORAGE { t.Errorf("Incorrect error returned for invalid storage: (%v)", e) }

    _, e = d.Create(ResourceType("crdt:gset"), "file", "invalid")
    if e == nil { t.Error("Invalid resource creation returned no error!") }
    if e != E_UNKNOWN_CRYPTO { t.Errorf("Incorrect error returned for invalid crypto: (%v)", e ) }
}

func Test_Database_Attach(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewSetResourceType(d, GROWONLYSET_RESOURCE_TYPE, NewGSetResource)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorage(NewFileStore("/tmp/crdb-test")); e != nil {
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

    qResource, e := d.Resolve(reference)
    if e != nil { t.Errorf("Failed to resolve reference: %v", e) }
    if qResource.Id() != resource.Id() { t.Error("ResourceId check failed!") }
}

func Test_Database_Attach_Invalid(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewSetResourceType(d, GROWONLYSET_RESOURCE_TYPE, NewGSetResource)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorage(NewFileStore("/tmp/crdb-test")); e != nil {
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
    if e != E_UNKNOWN_RESOURCE { t.Errorf("Incorrect error returned for invalid resource: %v", e) }

    _, e = d.Attach(resource.Id(), ResourceKey(""))
    if e == nil { t.Error("Invalid attach call returned no error!") }
    if e != E_INVALID_KEY { t.Errorf("Incorrect error returned for invalid key: %v", e) }
}

func Test_Database_Detach(t *testing.T) {
    d := NewDatabase()

    if e := d.RegisterType(NewSetResourceType(d, GROWONLYSET_RESOURCE_TYPE, NewGSetResource)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorage(NewFileStore("/tmp/crdb-test")); e != nil {
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

    if e := d.RegisterType(NewSetResourceType(d, GROWONLYSET_RESOURCE_TYPE, NewGSetResource)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorage(NewFileStore("/tmp/crdb-test")); e != nil {
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

    if e := d.RegisterType(NewSetResourceType(d, GROWONLYSET_RESOURCE_TYPE, NewGSetResource)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorage(NewFileStore("/tmp/crdb-test")); e != nil {
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
    if e := d.RegisterType(NewSetResourceType(d, GROWONLYSET_RESOURCE_TYPE, NewGSetResource)); e != nil {
        t.Errorf("Failed to register type: %v", e)
    }

    if e := d.RegisterStorage(NewFileStore("/tmp/crdb-test")); e != nil {
        t.Errorf("Failed to register storage: %v", e)
    }

    if e := d.RegisterCryptoMethod(aes); e != nil {
        t.Errorf("Failed to register crypto method: %v", e)
    }

    reference, e = d.Attach(resource.Id(), resource.Key())
    if e != nil { t.Errorf("Failed to attach resource: %v", e) }
}

