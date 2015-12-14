/*
 * Copyright (c) 2015 Tom Swindell (t.swindell@rubyx.co.uk)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/tswindell/go-crdt/db"
)

var (
    hostport = flag.String("hostport", "127.0.0.1:9600", "Database service host/port.")
)

type Command interface {
    RespondTo(cmd string) bool
    Execute(*crdb.Client)
}

type CRDBCommandListener struct{}

func (d *CRDBCommandListener) RespondTo(cmd string) bool {
    return cmd == "create" || cmd == "attach" || cmd == "detach" || cmd == "commit" || cmd == "list"
}

func (d *CRDBCommandListener) Execute(client *crdb.Client) {
    switch flag.Arg(0) {
    case "create": d.DoCreate(client)
    case "attach": d.DoAttach(client)
    case "detach": d.DoDetach(client)
    case "commit": d.DoCommit(client)
    case  "list": d.DoListTypes(client)
    }
}

func (d *CRDBCommandListener) DoCreate(client *crdb.Client) {
    if flag.NArg() < 4 {
        fmt.Fprintf(os.Stderr, "Usage: crdb-tool create <ResourceType> <StorageId> <CryptoTypeId>\n")
        os.Exit(1)
    }
    resourceId, resourceKey, e := client.Create(crdb.ResourceType(flag.Arg(1)), flag.Arg(2), flag.Arg(3))
    if e != nil {
        fmt.Fprintf(os.Stderr, "Error: Failed to execute create: %v\n", e)
        os.Exit(1)
    }

    fmt.Printf("ResourceId:%s\nResourceKey:%s\n", resourceId, resourceKey)
}

func (d *CRDBCommandListener) DoAttach(client *crdb.Client) {
    if flag.NArg() < 3 {
        fmt.Fprintf(os.Stderr, "Usage: crdb-tool attach <ResourceId> <ResourceKey>\n")
        os.Exit(1)
    }
    referenceId, e := client.Attach(crdb.ResourceId(flag.Arg(1)), crdb.ResourceKey(flag.Arg(2)))
    if e != nil {
        fmt.Fprintf(os.Stderr, "Error: Failed to execute attach: %v\n", e)
        os.Exit(1)
    }

    fmt.Printf("ReferenceId:%s\n", referenceId)
}

func (d *CRDBCommandListener) DoDetach(client *crdb.Client) {
    if flag.NArg() < 2 {
        fmt.Fprintf(os.Stderr, "Usage: crdb-tool detach <ReferenceId>\n")
        os.Exit(1)
    }

    if e := client.Detach(crdb.ReferenceId(flag.Arg(1))); e != nil {
        fmt.Fprintf(os.Stderr, "Error: Failed to execute detach: %v\n", e)
        os.Exit(1)
    }
}

func (d *CRDBCommandListener) DoCommit(client *crdb.Client) {
    if flag.NArg() < 2 {
        fmt.Fprintf(os.Stderr, "Usage: crdb-tool commit <ReferenceId>\n")
        os.Exit(1)
    }

    if e := client.Commit(crdb.ReferenceId(flag.Arg(1))); e != nil {
        fmt.Fprintf(os.Stderr, "Error: Failed to execute commit: %v\n", e)
        os.Exit(1)
    }
}

func (d *CRDBCommandListener) DoListTypes(client *crdb.Client) {
    if flag.NArg() < 2 {
        fmt.Fprintf(os.Stderr, "Usage: crdb-tool list <datatype|storage|crypto>\n")
        os.Exit(1)
    }

    fmt.Println("CRDT Supported Types:")
    var resourceTypes []string

    switch flag.Arg(1) {
    case "datatype":
        resourceTypes, _ = client.SupportedTypes()
    case "storage":
        resourceTypes, _ = client.SupportedStorageTypes()
    case "crypto":
        resourceTypes, _ = client.SupportedCryptoMethods()
    default:
        fmt.Fprintf(os.Stderr, "Usage: crdb-tool list <datatype|storage|crypto>\n")
        os.Exit(1)
    }

    for _, resourceType := range resourceTypes {
        var isSupportedCheck bool

        switch flag.Arg(1) {
        case "datatype":
            isSupportedCheck, _ = client.IsSupportedType(resourceType)
        case "storage":
            isSupportedCheck, _ = client.IsSupportedStorageType(resourceType)
        case "crypto":
            isSupportedCheck, _ = client.IsSupportedCryptoMethod(resourceType)
        }

        fmt.Printf("  %s - %v\n", resourceType, isSupportedCheck)
    }
    fmt.Println("")
}

func main() {
    flag.Parse()

    usage := `
    crdt-tool <commmand> [parameters ...]

     - OR -

    crdt-tool <data-type> <command> [parameters ...]

    Commands:
        create - Create a new resource.
        attach - Attach to resource and get reference.
        detach - Detach from resource and GC data.
        commit - Write modifications to persistent storage.
          list - List datatypes, storage types and crypto types.

`

    commands := make([]Command, 0)

    commands = append(commands, &CRDBCommandListener{})
    commands = append(commands, &CRDBGSetCommandListener{})
    commands = append(commands, &TwoPhaseSetCommandListener{})

    client := crdb.NewClient()
    if e := client.ConnectToHost(*hostport); e != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", e)
        os.Exit(1)
    }

    if flag.NArg() == 0 {
        fmt.Fprintf(os.Stderr, usage)
        fmt.Fprintf(os.Stderr, "Error: No command specified.\n")
        os.Exit(1)
    }

    var command Command
    for _, v := range commands {
        if v.RespondTo(flag.Arg(0)) {
            command = v
            break
        }
    }

    if command == nil {
        fmt.Fprintf(os.Stderr, usage)
        fmt.Fprintf(os.Stderr, "Error: Unknown command: %s\n", flag.Arg(0))
        os.Exit(1)
    }

    command.Execute(client)
    os.Exit(0)
}

