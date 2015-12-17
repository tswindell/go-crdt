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

