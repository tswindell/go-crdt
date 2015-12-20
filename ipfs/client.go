package ipfs

import (
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"

    "github.com/ipfs/go-ipfs/notifications"
    "github.com/ipfs/go-ipfs/p2p/peer"

    mh       "github.com/jbenet/go-multihash"
    base58   "github.com/jbenet/go-base58"

    cmd      "github.com/ipfs/go-ipfs/commands"
    files    "github.com/ipfs/go-ipfs/commands/files"
    http     "github.com/ipfs/go-ipfs/commands/http"
    commands "github.com/ipfs/go-ipfs/core/commands"
)

type Client struct {
    http.Client

    NodeInfo NodeId
}

type NodeId struct {
                 ID string
          PublicKey string
          Addresses []string
       AgentVersion string
    ProtocolVersion string
}

type QueryMessage struct {
           ID string
         Type int
    Responses []*peer.PeerInfo
        Extra string
}

type IpnsResponse struct {
    Path string
}

func NewClient(hostport string) *Client {
    d := new(Client)
    d.Client = http.NewClient(hostport)
    return d
}

func (d *Client) Connect() error {
    nodeInfo, e := d.Id()
    if e != nil { return e }

    d.NodeInfo = nodeInfo

    return nil
}

func (d *Client) IsConnected() bool {
    return len(d.NodeInfo.ID) > 0
}

func Multihash(data []byte) string {
    h, e := mh.Sum(data, mh.SHA2_256, -1)
    if e != nil { return "" }
    return base58.Encode(h)
}

func (d *Client) Get(path string) ([]byte, error) {
    return d.DoBasicRequest([]string{"cat", path}, nil, commands.CatCmd)
}

func (d *Client) Id() (NodeId, error) {
    var result NodeId

    decoder, e := d.DoJSONRequest([]string{"id"}, nil, commands.IDCmd)
    if e != nil { return result, e }
    decoder.Decode(&result)

    return result, nil
}

func (d *Client) ObjectGet(mh string) (commands.Node, error) {
    var result commands.Node

    decoder, e := d.DoJSONRequest([]string{"object", "get", mh}, nil, commands.ObjectCmd)
    if e != nil { return result, e }
    decoder.Decode(&result)

    return result, nil
}

func (d *Client) ObjectAddLink(mh string, name string, link string) (commands.Object, error) {
    var result commands.Object
    decoder, e := d.DoJSONRequest([]string{"object", "patch", mh},
                                  []string{"add-link", name, link},
                                  commands.ObjectCmd.Subcommands["patch"])
    if e != nil { return result, e }
    decoder.Decode(&result)
    return result, nil
}

func (d *Client) ObjectPutData(data []byte) (commands.Object, error) {
    return d.ObjectPutString(string(data))
}

func (d *Client) ObjectPutString(data string) (commands.Object, error) {
    node := commands.Node{Data: data}
    return d.ObjectPut(node)
}

func (d *Client) ObjectPut(node commands.Node) (commands.Object, error) {
    var result commands.Object

    path := []string{"object", "put"}
    opt, e := commands.ObjectCmd.GetOptions(nil)
    if e != nil { return result, e }

    nodeReader, nodeWriter := io.Pipe()

    file := files.NewReaderFile("", "", nodeReader, nil)
    dirw := files.NewSliceFile("", "", []files.File{file})

    go func() {
        encoder := json.NewEncoder(nodeWriter)
        encoder.Encode(node)
        nodeWriter.Close()
    }()

    req, e := cmd.NewRequest(path, nil, nil, dirw, commands.ObjectCmd, opt)
    if e != nil { return result, e }

    res, e := d.Send(req)
    if e != nil { return result, e }

    reader, e := res.Reader()
    if e != nil { return result, e }

    decoder := json.NewDecoder(reader)
    decoder.Decode(&result)

    return result, nil
}

func (d *Client) FindProvs(mh string, ch chan *peer.PeerInfo) error {
    path   := []string{"dht", "findprovs"}
    c := commands.DhtCmd.Subcommands["findprovs"]
    opt, e := c.GetOptions(nil)
    if e != nil { return e }

    req, e := cmd.NewRequest(path, nil, []string{mh}, nil, c, opt)
    if e != nil { return e }

    res, e := d.Send(req)
    if e != nil {
        LogError("Failed to send request: %v", e)
        d.NodeInfo = NodeId{}
        return e
    }

    reader, e := res.Reader()
    if e != nil { return e }
    defer res.Close()

    go func() {
        decoder := json.NewDecoder(reader)

        for decoder.More() {
            var mesg QueryMessage

            decoder.Decode(&mesg)
            if mesg.Type == int(notifications.PeerResponse) {
                for _, peer := range mesg.Responses { ch<- peer }
            }

            if mesg.Type == int(notifications.FinalPeer) { break }
            if mesg.Type == int(notifications.QueryError) { break }
        }

        LogInfo("Closing findprovs channel")
        res.Close()
        close(ch)
    }()

    return nil
}

func (d *Client) NamePublish(mh string) (commands.IpnsEntry, error) {
    var result commands.IpnsEntry

    decoder, e := d.DoJSONRequest([]string{"name", "publish", mh}, nil, commands.NameCmd)
    if e != nil { return result, e }
    decoder.Decode(&result)

    return result, nil
}

func (d *Client) NameResolve(mh string) (IpnsResponse, error) {
    var result IpnsResponse

    decoder, e := d.DoJSONRequest([]string{"name", "resolve", mh}, nil, commands.NameCmd)
    if e != nil { return result, e }
    decoder.Decode(&result)

    return result, nil
}

func (d *Client) DoReaderRequest(path []string, args []string, c *cmd.Command) (io.Reader, error) {
    opt, e := c.GetOptions(nil)
    if e != nil { return nil, e }

    req, e := cmd.NewRequest(path,  // Path
                             nil,   // options
                             args,  // args
                             nil,   // files.File
                             c,     // command
                             opt)   // options (defaults)
    if e != nil { return nil, e }

    res, e := d.Send(req)
    if e != nil {
        LogError("Failed to send request: %v", e)
        d.NodeInfo = NodeId{}
        return nil, e
    }

    if res.Error() != nil {
        LogError("Received error response: %s", res.Error().Message)
        return nil, fmt.Errorf(res.Error().Message)
    }

    defer res.Close()

    return res.Reader()
}

func (d *Client) DoBasicRequest(path []string, args []string, c *cmd.Command) ([]byte, error) {
    reader, e := d.DoReaderRequest(path, args, c)
    if e != nil { return nil, e }
    return ioutil.ReadAll(reader)
}

func (d *Client) DoJSONRequest(path []string, args []string, c *cmd.Command) (*json.Decoder, error) {
    reader, e := d.DoReaderRequest(path, args, c)
    if e != nil { return nil, e }
    return json.NewDecoder(reader), nil
}

