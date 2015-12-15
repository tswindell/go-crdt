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
    "strings"

    "code.google.com/p/go-uuid/uuid"
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
)

// The ResourceId type is the representation of a resources' identifier.
type ResourceId string

func NewResourceId(storageId, resourceId string) ResourceId {
    return ResourceId(storageId + ":" + resourceId)
}

// TODO: Consider renaming GetStorageId
func (d ResourceId) GetStorageId() string {
    return strings.SplitN(string(d), ":", 2)[0]
}

// TODO: Consider renaming GetId
func (d ResourceId) GetId() string {
    return strings.SplitN(string(d), ":", 2)[1]
}

func (d ResourceId) IsValid() bool {
    parts := strings.SplitN(string(d), ":", 2)
    return len(parts[0]) > 0 && len(parts[1]) > 0
}

// The ReferenceId type is the representation of a reference to a resource.
type ReferenceId string

func (d ReferenceId) IsValid() bool {
    return len(string(d)) > 0
}

// The ResourceKey type is the representation of a resources' encryption key.
type ResourceKey string

// Create a new ResourceKey instance
func NewResourceKey(typeId string, data []byte) ResourceKey {
    return ResourceKey(typeId + ":" + base64.StdEncoding.EncodeToString(data))
}

// Get the cryptographic method type id from key.
func (d ResourceKey) TypeId() string {
    return strings.SplitN(string(d), ":", 2)[0]
}

// Get the raw key data from key.
func (d ResourceKey) KeyData() []byte {
    keydata, _ := base64.StdEncoding.DecodeString(strings.SplitN(string(d), ":", 2)[1])
    return keydata
}

func (d ResourceKey) IsValid() bool {
    parts := strings.SplitN(string(d), ":", 2)
    return len(parts[0]) > 0
}

// The ResourceType type is the representation of a resources' data type.
type ResourceType string

func (d ResourceType) IsValid() bool {
    return len(string(d)) > 0
}

// The Resource interface defines the API that all resource types must provide.
type Resource interface {
    Id() ResourceId
    Key() ResourceKey
    TypeId() ResourceType

    Serialize(*bytes.Buffer) error
}

// The ResourceFactory interface defines the API that a resource type must
// provide.
type ResourceFactory interface {
    TypeId() ResourceType

    Create(ResourceId, ResourceKey) Resource
    Restore(ResourceId, ResourceKey, *bytes.Buffer) (Resource, error)
}

// The ResourceDatastore type
type ResourceDatastore map[ResourceId]Resource

// The ResourceTypeRegistry type
type ResourceTypeRegistry map[ResourceType]ResourceFactory

// The ReferenceTable type
type ReferenceTable map[ReferenceId]Resource

// The Datastore interface type defines the interface that persistent backing
// stores must implement.
type Datastore interface {
    TypeId() string

    HasResource(ResourceId) bool

    GetData(ResourceId) ([]byte, error)
    SetData(ResourceId, []byte) error
}

// The CryptoMethod interface type defines the interface that crypto methods
// must implement.
type CryptoMethod interface {
    TypeId() string

    GenerateKey() ResourceKey

    Encrypt(ResourceKey, []byte) ([]byte, error)
    Decrypt(ResourceKey, []byte) ([]byte, error)
}

// The Database type
type Database struct {
    datatypes  ResourceTypeRegistry
    datastore  ResourceDatastore
    references ReferenceTable

    crypto     map[string]CryptoMethod
    stores     map[string]Datastore
}

// The NewDatabase() function returns a newly created database instance.
func NewDatabase() *Database {
    d := new(Database)
    d.datatypes  = make(ResourceTypeRegistry)
    d.datastore  = make(ResourceDatastore)
    d.references = make(ReferenceTable)
    d.crypto     = make(map[string]CryptoMethod)
    d.stores     = make(map[string]Datastore)
    return d
}

// The RegisterType() function registers a new resource type factory within this
// instance.
func (d *Database) RegisterType(factory ResourceFactory) error {
    _, ok := d.datatypes[factory.TypeId()]
    if ok { return fmt.Errorf("Already registered datatype: %s", factory.TypeId()) }
    d.datatypes[factory.TypeId()] = factory
    return nil
}

// The RegisterCryptoMethod() instance method registers a cryptographic plugin
// with this database.
func (d *Database) RegisterCryptoMethod(method CryptoMethod) error {
    _, ok := d.crypto[method.TypeId()]
    if ok { return fmt.Errorf("Already registered crypto method: %s", method.TypeId()) }
    d.crypto[method.TypeId()] = method
    return nil
}

// The RegisterDatastore() instance method registers a datastore plugin with
// this database.
func (d *Database) RegisterStorageType(store Datastore) error {
    _, ok := d.stores[store.TypeId()]
    if ok { return fmt.Errorf("Already registered storage: %s", store.TypeId()) }
    d.stores[store.TypeId()] = store
    return nil
}

func (d *Database) StoreAll() error {
    for k, v := range d.datastore {
        k.GetStorageId()
        v.Key()
    }

    return nil
}

// The Create() database method creates a new resource from the specified parameters.
func (d *Database) Create(resourceType ResourceType, storageId string, cryptoId string) (Resource, error) {
    if !resourceType.IsValid() { return nil, E_INVALID_TYPE }

    factory, ok := d.datatypes[resourceType]
    if !ok { return nil, E_UNKNOWN_TYPE }

    // TODO: Replace with actual memcache "store" type.
    if storageId != "tmpfs" {
        _, ok = d.stores[storageId]
        if !ok { return nil, E_UNKNOWN_STORAGE }
    }

    crypto, ok := d.crypto[cryptoId]
    if !ok { return nil, E_UNKNOWN_CRYPTO }

    resourceId  := ResourceId(storageId + "://" + GenerateUUID())
    resourceKey := crypto.GenerateKey()

    resource := factory.Create(resourceId, resourceKey)
    d.datastore[resourceId] = resource

    return resource, nil
}

// The Attach() database method obtains a reference to a resource in the database.
func (d *Database) Attach(resourceId ResourceId, resourceKey ResourceKey) (ReferenceId, error) {
    resource, ok := d.datastore[resourceId]

    if !ok {
        var e error
        resource, e = d.Restore(resourceId, resourceKey)
        if e != nil { return ReferenceId(""), e }
    }

    if resource.Key() != resourceKey { return ReferenceId(""), E_INVALID_KEY }

    referenceId := ReferenceId(GenerateUUID())
    d.references[referenceId] = resource

    return referenceId, nil
}

// The Detach() database method removes a reference to a resource in the database.
func (d *Database) Detach(referenceId ReferenceId) error {
    if _, ok := d.references[referenceId]; !ok { return E_UNKNOWN_REFERENCE }
    delete(d.references, referenceId)
    return nil
}

// The Commit() database method commits a resource by reference to persistent
// storage.
func (d *Database) Commit(referenceId ReferenceId) error {
    resource, ok := d.references[referenceId]
    if !ok { return E_UNKNOWN_REFERENCE }

    storage, ok := d.stores[resource.Id().GetStorageId()]
    if !ok { return E_INVALID_STORAGE }

    crypto, ok := d.crypto[resource.Key().TypeId()]
    if !ok { return E_INVALID_CRYPTO }

    buff := bytes.Buffer{}
    buff.WriteString(string(resource.TypeId()))
    buff.WriteByte(byte(0x00))
    if e := resource.Serialize(&buff); e != nil { return e }

    data, e := crypto.Encrypt(resource.Key(), buff.Bytes())
    if e != nil { return e }

    e = storage.SetData(resource.Id(), data)
    if e != nil { return e }
    return nil
}

// The Restore() database method restores a resource from persistent storage.
func (d *Database) Restore(resourceId ResourceId, resourceKey ResourceKey) (Resource, error) {
    storage, ok := d.stores[resourceId.GetStorageId()]
    if !ok { return nil, E_UNKNOWN_RESOURCE }

    crypto, ok := d.crypto[resourceKey.TypeId()]
    if !ok { return nil, E_INVALID_KEY }

    data, e := storage.GetData(resourceId)
    if e != nil { return nil, e }

    data, e = crypto.Decrypt(resourceKey, data)
    if e != nil { return nil, e }

    buff := bytes.NewBuffer(data)
    typeString, e := buff.ReadString(byte(0x00))
    if e != nil { return nil, e }

    resourceType := ResourceType(typeString[:len(typeString) - 1])
    factory, ok := d.datatypes[resourceType]
    if !ok { return nil, E_UNKNOWN_TYPE }

    resource, e := factory.Restore(resourceId, resourceKey, buff)
    if e != nil { return nil, e }

    return resource, nil
}

// The SupportedTypes() database method returns a list of types that this
// database instance provides.
func (d *Database) SupportedTypes() []ResourceType {
    results := make([]ResourceType, 0)
    for k, _ := range d.datatypes {
        results = append(results, k)
    }
    return results
}

// The IsSupportedType() method returns whether a specific ResourceType is
// supported in this database.
func (d *Database) IsSupportedType(resourceType ResourceType) bool {
    _, ok := d.datatypes[resourceType]
    return ok
}

// The SupportedCryptoMethods() method returns a list of registered crypto
// methods supported in this database.
func (d *Database) SupportedCryptoMethods() []string {
    results := make([]string, 0)
    for k, _ := range d.crypto {
        results = append(results, k)
    }
    return results
}

// The IsSupportedCryptoMethod() method returns whether the supplied crypto
// method is supported in this database.
func (d *Database) IsSupportedCryptoMethod(crypto string) bool {
    _, ok := d.crypto[crypto]
    return ok
}

// The SupportedDatastores() method returns a list of registered storage
// backends.
func (d *Database) SupportedStorageTypes() []string {
    results := make([]string, 0)
    for k, _ := range d.stores {
        results = append(results, k)
    }
    return results
}

// The IsSupportedDatastore() method returns true if the supplied storage
// backend is supported in this database.
func (d *Database) IsSupportedStorageType(store string) bool {
    _, ok := d.stores[store]
    return ok
}

// The Resolve() database method resolves a ReferenceId to a ResourceId.
func (d *Database) Resolve(referenceId ReferenceId) (ResourceId, error) {
    res, ok := d.references[referenceId]
    if !ok { return ResourceId(""), E_INVALID_REFERENCE }
    return res.Id(), nil
}

// The GenerateUUID() function
func GenerateUUID() string {
    return uuid.New()
}

