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
    "fmt"
)

var (
    E_UNKNOWN_TYPE          = fmt.Errorf("crdt:unknown-resource-type")
    E_UNKNOWN_RESOURCE      = fmt.Errorf("crdt:unknown-resource-id")
    E_UNKNOWN_REFERENCE     = fmt.Errorf("crdt:unknown-reference-id")
    E_UNKNOWN_STORAGE       = fmt.Errorf("crdt:unknown-storage-type")
    E_UNKNOWN_CRYPTO        = fmt.Errorf("crdt:unknown-crypto-method")
    E_INVALID_TYPE          = fmt.Errorf("crdt:invalid-resource-type")
    E_INVALID_RESOURCE      = fmt.Errorf("crdt:invalid-resource-id")
    E_INVALID_KEY           = fmt.Errorf("crdt:invalid-resource-key")
    E_INVALID_REFERENCE     = fmt.Errorf("crdt:invalid-reference")
    E_INVALID_CRYPTO        = fmt.Errorf("crdt:invalid-crypto-id")
    E_INVALID_STORAGE       = fmt.Errorf("crdt:invalid-storage-id")
    E_INVALID_RESOURCE_DATA = fmt.Errorf("crdt:invalid-resource-data")
    E_TYPE_MISMATCH         = fmt.Errorf("crdt:resource-type-mismatch")
)


// The ResourceFactory interface defines the API that a resource type must
// provide.
type ResourceFactory interface {
    TypeId() ResourceType

    Create(ResourceId, ResourceKey) Resource

    Equals(Resource, Resource) (bool, error)

    Merge(Resource, Resource) error

    Clone(Resource) (Resource, error)

    Restore(ResourceId, ResourceKey, *bytes.Buffer) (Resource, error)
}


// The ResourceTypeRegistry type
type ResourceTypeRegistry ThreadSafeMap

func (d *ResourceTypeRegistry) AddFactory(factory ResourceFactory) bool {
    return ThreadSafeMap(*d).Insert(factory.TypeId(), factory)
}

func (d *ResourceTypeRegistry) GetFactory(typeId ResourceType) ResourceFactory {
    v := ThreadSafeMap(*d).GetValue(typeId)
    if v == nil { return nil }
    return v.(ResourceFactory)
}

func (d *ResourceTypeRegistry) List() []ResourceType {
    results := make([]ResourceType, 0)
    for _, v := range ThreadSafeMap(*d).Keys() { results = append(results, v.(ResourceType)) }
    return results
}


// The ResourceDatastore type
type ResourceDatastore ThreadSafeMap

func (d *ResourceDatastore) Add(resource Resource) bool {
    return ThreadSafeMap(*d).Insert(resource.Id(), resource)
}

func (d *ResourceDatastore) Get(resourceId ResourceId) Resource {
    v := ThreadSafeMap(*d).GetValue(resourceId)
    if v == nil { return nil }
    return v.(Resource)
}


// The ReferenceTable type
type ReferenceTable ThreadSafeMap

func (d *ReferenceTable) Create(resource Resource) ReferenceId {
    referenceId := ReferenceId(GenerateUUID())
    ThreadSafeMap(*d).Insert(referenceId, resource.Id())
    return referenceId
}

func (d *ReferenceTable) Remove(referenceId ReferenceId) bool {
    return ThreadSafeMap(*d).Remove(referenceId)
}

func (d *ReferenceTable) Resolve(referenceId ReferenceId) ResourceId {
    v := ThreadSafeMap(*d).GetValue(referenceId)
    if v == nil { return ResourceId("") }
    return v.(ResourceId)
}


// The Datastore interface type defines the interface that persistent backing
// stores must implement.
type Storage interface {
    TypeId() string

    GenerateResourceId() (ResourceId, error)

    HasResource(ResourceId) bool

    GetData(ResourceId, ResourceKey, chan []byte) error
    SetData(ResourceId, ResourceKey, []byte) error
}

// The StorageDirectory type
type StorageDirectory ThreadSafeMap

func (d *StorageDirectory) AddStore(store Storage) bool {
    return ThreadSafeMap(*d).Insert(store.TypeId(), store)
}

func (d *StorageDirectory) GetStore(storageId string) Storage {
    v := ThreadSafeMap(*d).GetValue(storageId)
    if v == nil { return nil }
    return v.(Storage)
}

func (d *StorageDirectory) List() []string {
    results := make([]string, 0)
    for _, v := range ThreadSafeMap(*d).Keys() { results = append(results, v.(string)) }
    return results
}


// The CryptoMethod interface type defines the interface that crypto methods
// must implement.
type CryptoMethod interface {
    TypeId() string

    GenerateKey() ResourceKey

    Encrypt(ResourceKey, []byte) ([]byte, error)
    Decrypt(ResourceKey, []byte) ([]byte, error)
}

// The CryptoMethodDirectory type
type CryptoMethodDirectory ThreadSafeMap

func (d *CryptoMethodDirectory) AddMethod(method CryptoMethod) bool {
    return ThreadSafeMap(*d).Insert(method.TypeId(), method)
}

func (d *CryptoMethodDirectory) GetMethod(methodId string) CryptoMethod {
    v := ThreadSafeMap(*d).GetValue(methodId)
    if v == nil { return nil }
    return v.(CryptoMethod)
}

func (d *CryptoMethodDirectory) List() []string {
    results := make([]string, 0)
    for _, v := range ThreadSafeMap(*d).Keys() { results = append(results, v.(string)) }
    return results
}


// The Database type
type Database struct {
    datatypes  ResourceTypeRegistry
    datastore  ResourceDatastore
    references ReferenceTable
    crypto     CryptoMethodDirectory
    storage    StorageDirectory
}

// The NewDatabase() function returns a newly created database instance.
func NewDatabase() *Database {
    d := new(Database)
    d.datatypes  = ResourceTypeRegistry(NewThreadSafeMap())
    d.datastore  = ResourceDatastore(NewThreadSafeMap())
    d.references = ReferenceTable(NewThreadSafeMap())
    d.storage    = StorageDirectory(NewThreadSafeMap())
    d.crypto     = CryptoMethodDirectory(NewThreadSafeMap())
    return d
}

// The RegisterType() function registers a new resource type factory within this
// instance.
func (d *Database) RegisterType(factory ResourceFactory) error {
    if !d.datatypes.AddFactory(factory) { return fmt.Errorf("Factory already registered: %s", factory.TypeId()) }
    return nil
}

// The RegisterStorage() instance method registers a datastore plugin with
// this database.
func (d *Database) RegisterStorage(store Storage) error {
    if !d.storage.AddStore(store) { return fmt.Errorf("Already registered storage: %s", store.TypeId()) }
    return nil
}

// The RegisterCryptoMethod() instance method registers a cryptographic plugin
// with this database.
func (d *Database) RegisterCryptoMethod(method CryptoMethod) error {
    if !d.crypto.AddMethod(method) { return fmt.Errorf("Already registered crypto method: %s", method.TypeId()) }
    return nil
}

// The Create() database method creates a new resource from the specified parameters.
func (d *Database) Create(resourceType ResourceType, storageId string, cryptoId string) (Resource, error) {
    if !resourceType.IsValid() { return nil, E_INVALID_TYPE }

    factory := d.datatypes.GetFactory(resourceType)
    if factory == nil { return nil, E_UNKNOWN_TYPE }

    storage := d.storage.GetStore(storageId)
    if storage == nil { return nil, E_UNKNOWN_STORAGE }

    crypto := d.crypto.GetMethod(cryptoId)
    if crypto == nil { return nil, E_UNKNOWN_CRYPTO }

    resourceId, e := storage.GenerateResourceId()
    if e != nil { return nil, E_INVALID_RESOURCE }

    resourceKey := crypto.GenerateKey()

    resource := factory.Create(resourceId, resourceKey)
    d.datastore.Add(resource)
    return resource, nil
}

// The Attach() database method obtains a reference to a resource in the database.
func (d *Database) Attach(resourceId ResourceId, resourceKey ResourceKey) (ReferenceId, error) {
    resource := d.datastore.Get(resourceId)

    if resource == nil {
        LogInfo("Resource not found in memory, attempting restore.")
        var e error
        resource, e = d.Restore(resourceId, resourceKey)
        if e != nil {
            LogInfo("Restore failed: %v", e)
            return ReferenceId(""), e
        }
    }

    if resource.Key() != resourceKey { return ReferenceId(""), E_INVALID_KEY }

    return d.references.Create(resource), nil
}

// The Detach() database method removes a reference to a resource in the database.
func (d *Database) Detach(referenceId ReferenceId) error {
    if !d.references.Remove(referenceId) { return E_INVALID_REFERENCE }
    return nil
}

// The Commit() database method commits a resource by reference to persistent
// storage.
func (d *Database) Commit(referenceId ReferenceId) error {
    resourceId := d.references.Resolve(referenceId)
    if !resourceId.IsValid() { return E_UNKNOWN_REFERENCE }

    resource := d.datastore.Get(resourceId)
    if resource == nil { return E_INVALID_RESOURCE }

    storage := d.storage.GetStore(resourceId.GetStorageId())
    if storage == nil { return E_INVALID_STORAGE }

    crypto := d.crypto.GetMethod(resource.Key().TypeId())
    if crypto == nil { return E_INVALID_CRYPTO }

    buff := bytes.Buffer{}
    buff.WriteString(string(resource.Type()))
    buff.WriteByte(byte(0x00))
    if e := resource.Serialize(&buff); e != nil { return e }

    data, e := crypto.Encrypt(resource.Key(), buff.Bytes())
    if e != nil { return e }

    e = storage.SetData(resource.Id(), resource.Key(), data)
    if e != nil { return e }
    return nil
}

// The Restore() database method restores a resource from persistent storage.
func (d *Database) Restore(resourceId ResourceId, resourceKey ResourceKey) (Resource, error) {
    var resource     Resource
    var resourceType ResourceType
    var factory      ResourceFactory
    var e            error

    storage := d.storage.GetStore(resourceId.GetStorageId())
    if storage == nil { return nil, E_UNKNOWN_RESOURCE }

    crypto := d.crypto.GetMethod(resourceKey.TypeId())
    if crypto == nil { return nil, E_INVALID_KEY }

    ch := make(chan []byte)

    go storage.GetData(resourceId, resourceKey, ch)

    for data := range ch {
        LogInfo("Decrypting stored data...")
        data, e = crypto.Decrypt(resourceKey, data)

        if e != nil {
            LogError("Decryption failed: %v", e)
            return nil, e
        }

        buff := bytes.NewBuffer(data)
        s, e := buff.ReadString(byte(0x00))
        if e != nil { return nil, e }

        iType := ResourceType(s[:len(s) - 1])

        LogInfo("Extracted resource type information: %s", iType)
        if resource == nil {
            resourceType = iType

            factory = d.datatypes.GetFactory(resourceType)
            if factory == nil {
                LogError("Failed to find factory for type: %s", resourceType)
                return nil, E_UNKNOWN_TYPE
            }

            LogInfo("Invoking factory restore method...")
            resource, e = factory.Restore(resourceId, resourceKey, buff)
            if e != nil || resource == nil {
                LogError("Factory restore failed: %v", e)
                return nil, e
            }
        } else {
            if iType != resourceType { return nil, E_TYPE_MISMATCH }

            // Acts as merge if resource is already initialized.
            e = resource.Deserialize(buff)
        }

    }

    LogInfo("Adding resource: %s", string(resource.Id()))
    d.datastore.Add(resource)
    return resource, nil
}

// The Equals() database method
func (d *Database) Equals(a ReferenceId, b ReferenceId) (bool, error) {
    aResource, e := d.Resolve(a)
    if e != nil { return false, e }

    bResource, e := d.Resolve(b)
    if e != nil { return false, e }

    // They can't be equal if they're not the same type.
    if aResource.Type() != bResource.Type() { return false, E_TYPE_MISMATCH }

    // Get resource factory for type.
    factory := d.datatypes.GetFactory(aResource.Type())
    if factory == nil { return false, E_INVALID_TYPE }

    return factory.Equals(aResource, bResource)
}

// The Merge() database method
func (d *Database) Merge(a ReferenceId, b ReferenceId) error {
    aResource, e := d.Resolve(a)
    if e != nil { return e }

    bResource, e := d.Resolve(b)
    if e != nil { return e }

    // Can't merge different types.
    if aResource.Type() != bResource.Type() { return E_TYPE_MISMATCH }

    // Get resource factory for type.
    factory := d.datatypes.GetFactory(aResource.Type())
    if factory == nil { return E_INVALID_TYPE }

    return factory.Merge(aResource, bResource)
}

// The Clone() database method
func (d *Database) Clone(referenceId ReferenceId) (Resource, error) {
    aResource, e := d.Resolve(referenceId)
    if e != nil { return nil, e }

    factory := d.datatypes.GetFactory(aResource.Type())
    if factory == nil { return nil, E_INVALID_TYPE }

    return factory.Clone(aResource)
}

// The SupportedTypes() database method returns a list of types that this
// database instance provides.
func (d *Database) SupportedTypes() []ResourceType {
    return d.datatypes.List()
}

// The IsSupportedType() method returns whether a specific ResourceType is
// supported in this database.
func (d *Database) IsSupportedType(resourceType ResourceType) bool {
    factory := d.datatypes.GetFactory(resourceType)
    return factory != nil
}

// The SupportedCryptoMethods() method returns a list of registered crypto
// methods supported in this database.
func (d *Database) SupportedCryptoMethods() []string {
    return d.crypto.List()
}

// The IsSupportedCryptoMethod() method returns whether the supplied crypto
// method is supported in this database.
func (d *Database) IsSupportedCryptoMethod(cryptoId string) bool {
    crypto := d.crypto.GetMethod(cryptoId)
    return crypto != nil
}

// The SupportedDatastores() method returns a list of registered storage
// backends.
func (d *Database) SupportedStorageTypes() []string {
    return d.storage.List()
}

// The IsSupportedDatastore() method returns true if the supplied storage
// backend is supported in this database.
func (d *Database) IsSupportedStorageType(store string) bool {
    storage := d.storage.GetStore(store)
    return storage != nil
}

// The Resolve() instance method
func (d *Database) Resolve(referenceId ReferenceId) (Resource, error) {
    resourceId := d.references.Resolve(referenceId)
    if !resourceId.IsValid() { return nil, E_INVALID_REFERENCE }

    resource := d.datastore.Get(resourceId)
    return resource, nil
}

