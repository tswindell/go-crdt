package ipfs

import (
    "testing"
)

func Test_Manifest(t *testing.T) {
    client   := NewClient("127.0.0.1:5001")
    if e := client.Connect(); e != nil {
       t.Fatal("Failed to connect IPFS client.")
    }

    manifest := NewManifest(client)
    if manifest == nil {
        t.Fatal("Failed to create manifest handler!")
    }

    for k, v := range manifest.Links {
        LogInfo("  Resource: %s - %s", k, v)
    }

    if e := manifest.Publish(); e != nil { 
        t.Fatal("Failed to publish manifest.")
    }
}

