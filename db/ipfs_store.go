package crdb

import (
    "crypto/rand"
    "encoding/base64"

    "github.com/tswindell/go-crdt/ipfs"
    "github.com/ipfs/go-ipfs/p2p/peer"
)

// The IPFSStore type
type IPFSStore struct {
    client   *ipfs.Client
    manifest *ipfs.Manifest
}

// Returns a new instance of IPFSStore type
func NewIPFSStore(hostport string) *IPFSStore {
    d := &IPFSStore{}
    d.client = ipfs.NewClient(hostport)
    if e := d.client.Connect(); e != nil {
        LogWarn("Failed to connect to IPFS daemon, is it running?")
    }
    d.manifest = ipfs.NewManifest(d.client)
    return d
}

// The TypeId() instance method
func (d *IPFSStore) TypeId() string { return "ipfs" }

func (d *IPFSStore) GenerateResourceId() (ResourceId, error) {
    objdata := make([]byte, 256)
    rand.Read(objdata)

    s, e := d.client.ObjectPutData(objdata)
    if e != nil { return ResourceId(""), e }

    return ResourceId(d.TypeId() + ":" + s.Hash), nil
}

// The HasResource() instance method
func (d *IPFSStore) HasResource(resourceId ResourceId) bool {
    //TODO: Should we return local cache presence through refs local?
    return resourceId.GetStorageId() == d.TypeId()
}

func GenerateLinkName(peerId peer.ID, id ResourceId, key ResourceKey) string {
    in := []byte(peerId.Pretty())
    in = append(in, []byte(id.GetId())...)
    in = append(in, key.KeyData()...)
    return ipfs.Multihash(in)
}

// The GetData() instance method
func (d *IPFSStore) GetData(id ResourceId, key ResourceKey, ch chan []byte) error {
    //TODO:
    //  - We need to add rules for large datasets
    //  - Each "resource" may have a "Previous" link, which points to
    //  another data resource segment?
    //
    // Here we go:
    //
    // for peer in findprovs(resourceId):
    //     linkname = sha256(peer.Id + resourceId + resourceKey)
    //     data := ipfs get /ipns/<peer.Id>/<linkname>
    //     ch<- data
    //
    // close(ch)
    //

    // Local retrieval:
    link := GenerateLinkName(peer.ID(d.client.NodeInfo.ID), id, key)
    LogInfo("Local Resource Link: %s", link)
    hash, ok := d.manifest.Links[link]

    if !ok {
        LogWarn("  Failed to find local data.")
    } else {
        LogInfo("  Link Target: %s", hash)

        node, e := d.client.ObjectGet(hash)
        if e == nil && len(node.Data) > 0 {
            v, e := base64.StdEncoding.DecodeString(node.Data)
            if e != nil {
                LogError("%v", e)
            } else {
                LogInfo("Sending local data")
                ch<- v
            }
        }
    }

    // Remotes retrieval:
    peers := make(chan *peer.PeerInfo)

    LogInfo("Initiating peer findprovs: %s", id.GetId())
    go d.client.FindProvs(id.GetId(), peers)

    for p := range peers {
        link := GenerateLinkName(p.ID, id, key)

        LogInfo("Found Peer Candidate: %s", p.ID.Pretty())
        LogInfo("  Attempting to resolve IPNS...")
        ipns, e := d.client.NameResolve(p.ID.Pretty())
        if e != nil { continue }

        LogInfo("  Resolved to: %s", ipns.Path)
        LogInfo("  Attempting to load manifest object...")
        manifest, e := d.client.ObjectGet(ipfs.StripHash(ipns.Path))
        if e != nil { continue }

        if manifest.Data != "crdt:Datastore" {
            LogWarn("  Returned object does not appear to be a valid manifest.")
            continue
        }

        for _, lnk := range manifest.Links {
            if lnk.Name != link { continue }

            LogInfo("  Found appropriate resource entry in manifest, loading data.")
            node, e := d.client.ObjectGet(lnk.Hash)
            if e != nil || len(node.Data) == 0 { continue }

            v, e := base64.StdEncoding.DecodeString(node.Data)
            if e != nil { continue }

            LogInfo("  Successfully decoded data, writing to output channel.")
            ch<- v
            break
        }

    }

    close(ch)

    return nil
}

// The SetData() instance method
func (d *IPFSStore) SetData(id ResourceId, key ResourceKey, data []byte) error {
    link := GenerateLinkName(peer.ID(d.client.NodeInfo.ID), id, key)

    h, e := d.client.ObjectPutString(base64.StdEncoding.EncodeToString(data))
    if e != nil { return e }

    if e := d.manifest.AddLink(link, ipfs.StripHash(h.Hash)); e != nil {
        return e
    }

    if e := d.manifest.Publish(); e != nil {
        LogError("Failed to publish manifest: %v", e)
    }

    return nil
}

