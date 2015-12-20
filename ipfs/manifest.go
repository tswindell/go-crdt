package ipfs

import (
    "fmt"
    "sync"
)

//TODO:
//  - Make manifests more resilient, don't rely on IPNS entry, keep hash values on disk as well.
//  - Could create a decentralized manifest backup/query/restore system.

type Manifest struct {
    sync.RWMutex
    Client

    PeerId string
    Hash   string // Current hash
    Links  map[string]string
}

func NewManifest(client *Client) *Manifest {
    d := &Manifest{}
    d.Client = *client

    if len(d.Client.NodeInfo.ID) == 0 { return nil }

    d.PeerId = d.Client.NodeInfo.ID

    if e := d.Resolve(); e != nil || len(d.Hash) == 0 {
        d.InitManifest()
    }

    if e := d.Refresh(); e != nil {
        d.InitManifest()
    }

    return d
}

func (d *Manifest) IsValid() bool {
    return d.Hash != ""
}

func (d *Manifest) InitManifest() error {
    d.Lock()
    defer d.Unlock()

    if d.Hash != "" {
        return fmt.Errorf("Can't initialize an already existing manifest!")
    }

    LogInfo("Initializing new IPFS manifest data structure.")
    obj, e := d.Client.ObjectPutString("crdt:Datastore")
    if e != nil { return e }

    d.Hash = StripHash(obj.Hash)
    return nil
}

func (d *Manifest) AddLink(name, link string) error {
    d.Lock()
    defer d.Unlock()

    LogInfo("Adding link: %s -> %s", name, link)
    obj, e := d.Client.ObjectAddLink(d.Hash, name, link)
    if e != nil { return e }
    if len(obj.Hash) == 0  { return fmt.Errorf("Invalid response") }

    LogInfo("New manifest hash: %s", obj.Hash)
    d.Hash = obj.Hash
    d.Links[name] = link
    return nil
}

func (d *Manifest) Refresh() error {
    d.Lock()
    defer d.Unlock()

    d.Links = make(map[string]string)

    node, e := d.Client.ObjectGet(d.Hash)
    if e != nil {
        d.Hash = ""
        return e
    }

    if node.Data != "crdt:Datastore" {
        d.Hash = ""
        return fmt.Errorf("Invalid manifest!")
    }

    LogInfo("Manifest: %s", d.Hash)
    for _, v := range node.Links {
        LogInfo("  - Link: %s -> %s", v.Name, v.Hash)
        d.Links[v.Name] = v.Hash
    }

    return nil
}

func (d *Manifest) Resolve() error {
    d.Lock()
    defer d.Unlock()

    LogInfo("Resolving: %s", d.PeerId)
    ipns, e := d.Client.NameResolve(d.PeerId)
    if e != nil { return e }
    if len(ipns.Path) == 0 { return fmt.Errorf("Empty IPNS response received.") }

    LogInfo("    Resolved %s -> %s", d.PeerId, ipns.Path)
    d.Hash = StripHash(ipns.Path)
    return nil
}

func (d *Manifest) Publish() error {
    d.RLock()
    defer d.RUnlock()

    LogInfo("Publishing manifest: %s", d.Hash)
    _, e := d.Client.NamePublish(d.Hash)
    return e
}

