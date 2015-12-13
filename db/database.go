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

    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "fmt"

    "code.google.com/p/go-uuid/uuid" 
)

var (
    E_UNKNOWN_TYPE      = fmt.Errorf("crdt:unknown-resource-type")
    E_UNKNOWN_RESOURCE  = fmt.Errorf("crdt:unknown-resource-id")
    E_INVALID_KEY       = fmt.Errorf("crdt:invalid-resource-key")
    E_INVALID_REFERENCE = fmt.Errorf("crdt:invalid-reference")
)

// The ResourceId type is the representation of a resources' identifier.
type ResourceId string

// The ReferenceId type is the representation of a reference to a resource.
type ReferenceId string

// The ResourceKey type is the representation of a resources' encryption key.
type ResourceKey []byte

// The ResourceType type is the representation of a resources' data type.
type ResourceType string

// The Resource interface defines the API that all resource types must provide.
type Resource interface {
    Id() ResourceId
    Key() ResourceKey
    Type() ResourceType

    Save() []byte
    Load([]byte) error
}

// The ResourceFactory interface defines the API that a resource type must
// provide.
type ResourceFactory interface {
    Type() ResourceType

    Create(ResourceId, ResourceKey) Resource
}

// The ResourceDatastore type
type ResourceDatastore map[ResourceId]Resource

// The ResourceTypeRegistry type
type ResourceTypeRegistry map[ResourceType]ResourceFactory

// The ReferenceTable type
type ReferenceTable map[ReferenceId]Resource

// The Database type
type Database struct {
    datatypes  ResourceTypeRegistry
    datastore  ResourceDatastore
    references ReferenceTable
}

// The NewDatabase() function returns a newly created database instance.
func NewDatabase() *Database {
    d := new(Database)
    d.datatypes  = make(ResourceTypeRegistry)
    d.datastore  = make(ResourceDatastore)
    d.references = make(ReferenceTable)
    return d
}

func (d *Database) RegisterType(factory ResourceFactory) {
    d.datatypes[factory.Type()] = factory
}

// The Create() database method creates a new resource from the specified parameters.
func (d *Database) Create(resourceType ResourceType) (ResourceId, ResourceKey, error) {
    factory, ok := d.datatypes[resourceType]
    if !ok { return ResourceId(""), ResourceKey(""), E_UNKNOWN_TYPE }

    resourceId  := ResourceId(GenerateUUID())
    resourceKey := ResourceKey(GenerateRandomKey())

    resource := factory.Create(resourceId, resourceKey)
    d.datastore[resourceId] = resource

    return resourceId, resourceKey, nil
}

// The Attach() database method obtains a reference to a resource in the database.
func (d *Database) Attach(resourceId ResourceId, resourceKey ResourceKey) (ReferenceId, error) {
    resource, ok := d.datastore[resourceId]
    if !ok { return ReferenceId(""), E_UNKNOWN_RESOURCE }

    if !bytes.Equal(resource.Key(), resourceKey) { return ReferenceId(""), E_INVALID_KEY }

    referenceId := ReferenceId(GenerateUUID())
    d.references[referenceId] = resource

    return referenceId, nil
}

// The Detach() database method removes a reference to a resource in the database.
func (d *Database) Detach(referenceId ReferenceId) error {
    if _, ok := d.references[referenceId]; !ok { return E_INVALID_REFERENCE }
    delete(d.references, referenceId)
    return nil
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

// The IsSupportedType() database method queries the database
func (d *Database) IsSupportedType(resourceType ResourceType) bool {
    _, ok := d.datatypes[resourceType]
    return ok
}

func (d *Database) Resolve(referenceId ReferenceId) (ResourceId, error) {
    res, ok := d.references[referenceId]
    if !ok { return ResourceId(""), E_INVALID_REFERENCE }
    return res.Id(), nil
}

// The GenerateUUID() function
func GenerateUUID() string {
    return uuid.New()
}

// The GenerateRandomKey() function
func GenerateRandomKey() string {
    raw := make([]byte, 512)
    rand.Read(raw)
    hash := sha256.New()
    hash.Write(raw[:])
    return hex.EncodeToString(hash.Sum(nil))
}

