package crdb

import (
    "bytes"
    "encoding/base64"
    "strings"
)


// The ResourceId type is the representation of a resource identifier.
type ResourceId  string

func NewResourceId(storageId, resourceId string) ResourceId {
    return ResourceId(storageId + ":" + resourceId)
}

func (d ResourceId) GetId() string {
    return strings.SplitN(string(d), ":", 2)[1]
}

func (d ResourceId) GetStorageId() string {
    return strings.SplitN(string(d), ":", 2)[0]
}

func (d ResourceId) IsValid() bool {
    parts := strings.SplitN(string(d), ":", 2)
    return len(parts) == 2 && len(parts[0]) > 0 && len(parts[1]) > 0
}


// The ResourceKey type is the representation of a resource encryption key.
type ResourceKey string

func NewResourceKey(typeId string, data []byte) ResourceKey {
    return ResourceKey(typeId + ":" + base64.StdEncoding.EncodeToString(data))
}

func (d ResourceKey) TypeId() string {
    return strings.SplitN(string(d), ":", 2)[0]
}

func (d ResourceKey) KeyData() []byte {
    keydata, _ := base64.StdEncoding.DecodeString(strings.SplitN(string(d), ":", 2)[1])
    return keydata
}

func (d ResourceKey) IsValid() bool {
    parts := strings.SplitN(string(d), ":", 2)
    return len(parts[0]) > 0
}


// The ReferenceId type is the representation of a resource reference identifier.
type ReferenceId string

func (d ReferenceId) IsValid() bool {
    return len(string(d)) > 0
}


// The ResourceType type is the representation of a resource datatype identifier.
type ResourceType string

func (d ResourceType) IsValid() bool {
    return len(string(d)) > 0
}


// The Resource interface defines the API that a resource type must implement.
type Resource interface {
    Id() ResourceId
    Key() ResourceKey
    Type() ResourceType

    Serialize(*bytes.Buffer) error
    Deserialize(*bytes.Buffer) error
}


// The ResourceBase type is a base class for resource types to include.
type ResourceBase struct {
    id       ResourceId
    key      ResourceKey
    datatype ResourceType
}

func (d *ResourceBase) Id() ResourceId { return d.id }

func (d *ResourceBase) Key() ResourceKey { return d.key }

func (d *ResourceBase) Type() ResourceType { return d.datatype }

