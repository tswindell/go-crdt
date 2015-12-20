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

    ch := make(chan []byte)
    if e := fs.GetData(ResourceId("file://0123456789ABCDEF"), ResourceKey(""), ch); e == nil {
        t.Error("Expected failure from GetData")
    }

    ch = make(chan []byte)
    if e := fs.GetData(ResourceId("file://0123456789ABCDEF"), ResourceKey(""), ch); e != E_UNKNOWN_RESOURCE {
        t.Errorf("Wrong error returned from bad GetData: %v", e)
    }

    in := make([]byte, 32)
    rand.Read(in)

    if e := fs.SetData(ResourceId("file://0123456789ABCDEF"), ResourceKey(""), in); e != nil {
        t.Errorf("Failed call to SetData with valid data: %v", e)
    }

    ch = make(chan []byte)
    go fs.GetData(ResourceId("file://0123456789ABCDEF"), ResourceKey(""), ch)

    out := <-ch

    if !bytes.Equal(in, out) {
        t.Error("Loaded data does not equal saved data!")
    }
}

