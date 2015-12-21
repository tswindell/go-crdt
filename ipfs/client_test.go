package ipfs

import (
    "testing"
    "fmt"

    "github.com/ipfs/go-ipfs/p2p/peer"

    commands "github.com/ipfs/go-ipfs/core/commands"
)

var (
  cl *Client
)

func initClient(t *testing.T) error {
    cl = NewClient("127.0.0.1:5001")
    return cl.Connect()
}

func Test_Multihash(t *testing.T) {
    mh := Multihash([]byte("Hello, world!"))
    if len(mh) == 0 { t.Error("Multihash has 0 length") }
}

func Test_Client_New(t *testing.T) {
    if e := initClient(t); e != nil {
        t.Fatal("Failed to connect.")
    }
}

func Test_Client_New_Invalid(t *testing.T) {
    client := NewClient("invalid")
    if e := client.Connect(); e == nil {
        t.Fatal("Connect returned no error with invalid data!")
    }
}

func Test_Client_Id(t *testing.T) {
    if e := initClient(t); e != nil { t.Fatal("Failed to connect") }

    node, e := cl.Id()
    if e != nil { t.Error(e) }

    fmt.Println(" Node ID:", node.ID)
    fmt.Println("Node Key:", node.PublicKey)
}

func Test_Client_ObjectGet(t *testing.T) {
    if e := initClient(t); e != nil { t.Fatal("Failed to connect") }

    // Get empty unixdir
    node, e := cl.ObjectGet("QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn")
    if e != nil { t.Error(e) }

    if len(node.Links) != 0 { t.Error("Links length != 0") }
    if len(node.Data) != 2 { t.Error("Data length != 2") }
}

func Test_Client_ObjectGet_Invalid(t *testing.T) {
    if e := initClient(t); e != nil { t.Fatal("Failed to connect") }

    _, e := cl.ObjectGet("invalid")
    if e == nil { t.Fatal("No error received with invalid data!") }
}

func Test_Client_ObjectPutData(t *testing.T) {
    if e := initClient(t); e != nil { t.Fatal("Failed to connect") }

    h, e := cl.ObjectPutData([]byte("Hello, world"))
    if e != nil { t.Error(e) }

    fmt.Println(" Hash:", h)
}

func Test_Client_ObjectPut(t *testing.T) {
}

func Test_Client_FindProvs(t *testing.T) {
    if e := initClient(t); e != nil { t.Error("Failed to connect") }

    ch := make(chan *peer.PeerInfo)

    // Everyone has a copy of the empty directory IPFS object.
    go cl.FindProvs("QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn", ch)

    for p := range ch {
        LogInfo("  Found peer: %s", p.ID.Pretty())
    }
}

func Test_Client_NamePublish(t *testing.T) {
    if e := initClient(t); e != nil { t.Fatal("Failed to connect") }

    _, e := cl.NamePublish("QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn")
    if e != nil { t.Fatal(e) }
}

func Test_Client_NameResolve(t *testing.T) {
    if e := initClient(t); e != nil { t.Fatal("Failed to connect") }

    r, e := cl.NameResolve(cl.PeerId)
    if e != nil { t.Fatal(e) }

    if len(r.Path) == 0 {
        t.Error("Path is zero length!")
    }
}

func Test_Client(t *testing.T) {
    client    := NewClient("127.0.0.1:5001")

    onError := func(err error) {
        t.Error(e.Error())
    }

    if e := client.Connect(); e != nil {
        t.Fatal("Failed to connect")
    }

    peerId, e := client.Id()
    fmt.Println("Peer ID: ", peerId.ID)

    obj, e := client.ObjectPut(commands.Node{Data: `Hello, world!`})
    if e != nil { onError(e) }
    fmt.Println("    Test Object: ", obj.Hash)

    obj, e = client.ObjectPutString("Hello, world!")
    if e != nil { onError(e) }

    node, e := client.ObjectGet(obj.Hash)
    if e != nil { onError(e) }
    fmt.Println(node.Data)

    name, e := client.NamePublish(obj.Hash)
    if e != nil { onError(e) }

    fmt.Println("New Published IPNS Record:")
    fmt.Println("   Name: ", name.Name)
    fmt.Println("  Value: ", name.Value)
    fmt.Println("")

    ch := make(chan *peer.PeerInfo)
    go client.FindProvs(obj.Hash, ch)

    for peer := range ch {
        fmt.Println("PeerId: ", peer.ID.Pretty())
    }
}

