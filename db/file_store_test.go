package crdb

import "testing"
import "bytes"
import "crypto/rand"
import "os"

const FILESTORE_TEST_PATH = "/tmp/crdb-fstore-test"

func Test_FileStore(t *testing.T) {
    os.RemoveAll(FILESTORE_TEST_PATH)

    fs := NewFileStore(FILESTORE_TEST_PATH)

    if fs.TypeId() != "file" { t.Errorf("Wrong TypeId returned, got: %s", fs.TypeId()) }

    if fs.HasResource(ResourceId("file://0123456789ABCDEF")) {
        t.Error("Expected false from invalid HasResource")
    }

    if _, e := fs.GetData(ResourceId("file://0123456789ABCDEF")); e == nil {
        t.Error("Expected failure from GetData")
    }

    if _, e := fs.GetData(ResourceId("file://0123456789ABCDEF")); e != E_UNKNOWN_RESOURCE {
        t.Errorf("Wrong error returned from bad GetData: %v", e)
    }

    in := make([]byte, 32)
    _, e := rand.Read(in)

    if e := fs.SetData(ResourceId("file://0123456789ABCDEF"), in); e != nil {
        t.Errorf("Failed call to SetData with valid data: %v", e)
    }

    out, e := fs.GetData(ResourceId("file://0123456789ABCDEF"))
    if e != nil {
        t.Errorf("Failed to call GetData with valid data: %v", e)
    }

    if !bytes.Equal(in, out) {
        t.Error("Loaded data does not equal saved data!")
    }
}

