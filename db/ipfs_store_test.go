package crdb

import "testing"
import "fmt"

import "bytes"
import "crypto/rand"

func Test_IPFSStore(t *testing.T) {
    fs := NewIPFSStore("127.0.0.1:5001")

    if !fs.HasResource(ResourceId("ipfs:0123456789ABCDEF")) {
        t.Error("Expected true from invalid HasResource")
    }

    in := make([]byte, 32)
    rand.Read(in)

    resourceId, e := fs.GenerateResourceId()
    if e != nil {
        t.Errorf("Failed to generate new resource id")
    }

    if e := fs.SetData(resourceId, ResourceKey("none:abcd"), in); e != nil {
        t.Errorf("Failed call to SetData with valid data: %v", e)
    }

    ch := make(chan []byte)
    go fs.GetData(resourceId, ResourceKey("none:abcd"), ch)

    out := <-ch

    fmt.Printf("%x\n", in)
    fmt.Printf("%x\n", out)

    if !bytes.Equal(in, out) {
        t.Error("Loaded data does not equal saved data!")
    }
}

