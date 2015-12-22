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

// The GenerateResourceId() instance method, generates some random data, and pins
// it in IPFS to act as a resource identifier in findprovs.
func (d *IPFSStore) GenerateResourceId() (ResourceId, error) {
    objdata := make([]byte, 256)
    rand.Read(objdata)

    //FIXME: This could probably be done better somewhere else ...
    s, e := d.client.ObjectPutData(objdata)
    if e != nil { return ResourceId(""), e }

    return ResourceId(d.TypeId() + ":" + s.Hash), nil
}

// The HasResource() instance method
func (d *IPFSStore) HasResource(resourceId ResourceId) bool {
    //TODO:
    //  - Should we return local cache presence through refs local?
    //    . Maybe we can add HasLocalResource(resourceId ResourceId) method. 
    return resourceId.GetStorageId() == d.TypeId()
}

// The GenerateLinkName() function, creates a determinate IPFS style base58
// encoded hash to use as a unique link name in a CRDT manifest node. This
// hash is unique to the peerId, so no two IPFS peers share the same link
// name. This adds a level of obfuscation to the manifest resource system
// to allow some level of data masking.
func GenerateLinkName(peerId string, id ResourceId, key ResourceKey) string {
    in := []byte(peerId)
    in = append(in, []byte(id.GetId())...)
    in = append(in, key.KeyData()...)
    LogInfo("GenerateLinkName: PeerId=%s, Id=%s, Key=%x :: Hash=%s",
                               peerId, id.GetId(), key.KeyData(), ipfs.Multihash(in))
    return ipfs.Multihash(in)
}

// The GetData() instance method, is the IPFS storage types implementation of the
// crdb.Storage interface.
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
    LogInfo("GetData: ResourceID=%s, ResourceKey=%x", id.GetId(), key.KeyData())

    // Local retrieval:
    link := GenerateLinkName(d.client.PeerId, id, key)
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
    pch := make(chan *peer.PeerInfo)
    go d.client.FindProvs(id.GetId(), pch)

    //FIXME: Work around for buggy client not handling multiple simultaneous
    //     IPFS queries.
    peers := make([]string, 0)
    for p := range pch {
        peers = append(peers, p.ID.Pretty())
    }

    for _, p := range peers {
        LogInfo("Tracking peer: %s", p)
        link := GenerateLinkName(p, id, key)

        LogInfo("Found Peer Candidate: %s", p)
        LogInfo("  Attempting to resolve IPNS...")
        ipns, e := d.client.NameResolve(p)
        if e != nil {
            LogError("  Failed to resolve IPNS: %v", e)
            continue
        }

        LogInfo("  Resolved to: %s", ipns.Path)
        LogInfo("  Downloading manifest object from peer...")
        parts := ipns.Path.Segments()
        manifest, e := d.client.ObjectGet(parts[len(parts)-1])
        if e != nil {
            LogError("  Failed to retrieve manifest: %v", e)
            continue
        }

        if manifest.Data != "crdt:Datastore" {
            LogError("  Returned object does not appear to be a valid manifest.")
            continue
        }

        //TODO: Could this be made more simple by retrieving IPNS sub-path?
        //          I.E. - /ipns/<PEER_ID>/<GENERATED_LINK_NAME>
        for _, lnk := range manifest.Links {
            if lnk.Name != link {
                LogInfo("Skipping link: %s != %s", lnk.Name, link)
                continue
            }

            LogInfo("  Found appropriate resource entry in manifest, loading data.")
            node, e := d.client.ObjectGet(lnk.Hash)
            if e != nil {
                LogError("  Failed to retrieve CRDT object data: %v", e)
                break
            }

            if len(node.Data) == 0 {
                LogError("  Received empty CRDT object data!")
                break
            }

            v, e := base64.StdEncoding.DecodeString(node.Data)
            if e != nil {
                LogError("  Failed to decode IPFS node data: %v", e)
                break
            }

            LogInfo("  Successfully decoded data, writing to output channel.")
            ch<- v

            LogInfo("  Downloading resource object.")
            if _, e := d.client.ObjectGet(id.GetId()); e != nil {
                LogError("  Failed to download resource object: %v", e)
                break
            }
            break
        }

    }

    close(ch)
    return nil
}

// The SetData() instance method, is the IPFS storage types implementation of
// the crdb Storage interface.
func (d *IPFSStore) SetData(id ResourceId, key ResourceKey, data []byte) error {
    link := GenerateLinkName(d.client.PeerId, id, key)

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

