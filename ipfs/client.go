package ipfs

import (
    "encoding/json"
    "fmt"
    "io"
    "time"

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
    hostport string

    PeerId string
}

func NewClient(hostport string) *Client {
    d := new(Client)
    d.hostport = hostport
    return d
}

func (d *Client) Connect() error {
    nodeInfo, e := d.Id()
    if e != nil { return e }

    d.PeerId = nodeInfo.ID

    return nil
}

func (d *Client) IsConnected() bool {
    return len(d.PeerId) > 0
}

func Multihash(data []byte) string {
    h, e := mh.Sum(data, mh.SHA2_256, -1)
    if e != nil { return "" }
    return base58.Encode(h)
}

func (d *Client) Id() (commands.IdOutput, error) {
    response, e := d.DoRequest([]string{"id"}, nil, commands.IDCmd)
    if e != nil { return commands.IdOutput{}, e }
    defer response.Close()

    value, ok := response.Output().(*commands.IdOutput)
    if !ok {
        return commands.IdOutput{}, fmt.Errorf("Failed to cast output object.")
    }
    return *value, nil
}

func (d *Client) ObjectGet(mh string) (commands.Node, error) {
    response, e := d.DoRequest([]string{"object", "get", mh}, nil,
                                        commands.ObjectCmd.Subcommands["get"])
    if e != nil { return commands.Node{}, e }
    defer response.Close()

    value, ok := response.Output().(*commands.Node)
    if !ok {
        return commands.Node{}, fmt.Errorf("Failed to cast output object.")
    }
    return *value, nil
}

func (d *Client) ObjectAddLink(mh string, name string, link string) (commands.Object, error) {
    response, e := d.DoRequest([]string{"object", "patch", mh},
                               []string{"add-link", name, link},
                               commands.ObjectCmd.Subcommands["patch"])
    if e != nil { return commands.Object{}, e }
    defer response.Close()

    value, ok := response.Output().(*commands.Object)
    if !ok {
        return commands.Object{}, fmt.Errorf("Failed to cast output object.")
    }
    return *value, nil
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

    client := http.NewClient(d.hostport)

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

    res, e := client.Send(req)
    if e != nil { return result, e }
    defer res.Close()

    reader, e := res.Reader()
    if e != nil { return result, e }

    decoder := json.NewDecoder(reader)
    decoder.Decode(&result)

    return result, nil
}

func (d *Client) FindProvs(mh string, ch chan *peer.PeerInfo) error {
    client := http.NewClient(d.hostport)

    defer close(ch)

    path   := []string{"dht", "findprovs"}
    c := commands.DhtCmd.Subcommands["findprovs"]
    opt, e := c.GetOptions(nil)
    if e != nil { return e }

    LogInfo("Searching for peers who provide: %s", mh)
    req, e := cmd.NewRequest(path, nil, []string{mh}, nil, c, opt)
    if e != nil { return e }

    res, e := client.Send(req)
    if e != nil {
        LogError("Failed to send request: %v", e)
        d.PeerId = ""
        return e
    }
    defer res.Close()

    if res.Error() != nil {
        LogError(res.Error().Message)
        return fmt.Errorf(res.Error().Message)
    }

    out, ok := res.Output().(<-chan interface{})
    if !ok {
        LogError("Failed to get output channel")
        return fmt.Errorf("Failed to get output channel")
    }

    for {
        var obj *notifications.QueryEvent
        var ok   bool

        select {
            case v := <-out:
                obj, ok = v.(*notifications.QueryEvent)

            case <-time.After(time.Second * 5):
                LogInfo("Request timeout occured, closing connection.")
                return nil
        }

        if !ok {
            LogError("Failed to cast QueryEvent")
            break
        }

        if obj.Type == notifications.Provider {
            for _, peer := range obj.Responses {
                LogInfo("Adding peer response for: %s", peer.ID.Pretty())
                ch<- peer
            }
        }
    }

    return nil
}

func (d *Client) NamePublish(mh string) (commands.IpnsEntry, error) {
    response, e := d.DoRequest([]string{"name", "publish", mh}, nil, commands.PublishCmd)
    if e != nil { return commands.IpnsEntry{}, e }
    defer response.Close()

    value, ok := response.Output().(*commands.IpnsEntry)
    if !ok {
        return commands.IpnsEntry{}, fmt.Errorf("Failed to cast output object.")
    }
    return *value, nil
}

func (d *Client) NameResolve(mh string) (commands.ResolvedPath, error) {
    response, e := d.DoRequest([]string{"name", "resolve", mh}, nil, commands.ResolveCmd)
    if e != nil { return commands.ResolvedPath{}, e }
    response.Close()

    value, ok := response.Output().(*commands.ResolvedPath)
    if !ok {
        return commands.ResolvedPath{}, fmt.Errorf("Failed to cast output object.")
    }
    return *value, nil
}

func (d *Client) DoRequest(path []string, args []string, c *cmd.Command) (cmd.Response, error) {
    client := http.NewClient(d.hostport)

    opt, e := c.GetOptions(nil)
    if e != nil { return nil, e }

    req, e := cmd.NewRequest(path,  // Path
                             nil,   // options
                             args,  // args
                             nil,   // files.File
                             c,     // command
                             opt)   // options (defaults)
    if e != nil { return nil, e }

    res, e := client.Send(req)
    if e != nil {
        LogError("Failed to send request: %v", e)
        d.PeerId = ""
        return nil, e
    }

    if res.Error() != nil {
        LogError("Received error response: %s", res.Error().Message)
        res.Close()
        return nil, fmt.Errorf(res.Error().Message)
    }

    return res, nil
}

