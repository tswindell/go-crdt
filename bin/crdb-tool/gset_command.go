package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/tswindell/go-crdt/db"
)

type CRDBGSetCommandListener struct {}

func (d *CRDBGSetCommandListener) RespondTo(cmd string) bool {
    return cmd == "crdt:gset"
}

func (d *CRDBGSetCommandListener) ShowUsage(usage string) {
    fmt.Fprintf(os.Stderr, "Usage: crdb-tool crdt:gset %s\n", usage)
}

func (d *CRDBGSetCommandListener) CheckNArg(count int, usage string) {
    if flag.NArg() < count {
        d.ShowUsage(usage)
        os.Exit(1)
    }
}

func (d *CRDBGSetCommandListener) CheckError(m string, e error) {
    if e != nil {
        fmt.Fprintf(os.Stderr, "Error: %s: %v\n", m, e)
        os.Exit(1)
    }
}

func (d *CRDBGSetCommandListener) DoList(client *crdb.Client) {
    d.CheckNArg(3, "list <ReferenceId>")

    ch, e := client.GSetClient.List(crdb.ReferenceId(flag.Arg(2)))
    d.CheckError("Failed to list set", e)

    for item := range ch {
        fmt.Println(item)
    }
}

func (d *CRDBGSetCommandListener) DoInsert(client *crdb.Client) {
    d.CheckNArg(4, "insert <ReferenceId> <OBJECT_DATA>")

    e := client.GSetClient.Insert(crdb.ReferenceId(flag.Arg(2)), []byte(flag.Arg(3)))
    d.CheckError("Failed to insert item", e)
}

func (d *CRDBGSetCommandListener) DoLength(client *crdb.Client) {
    d.CheckNArg(3, "length <ReferenceId>")

    length, e := client.GSetClient.Length(crdb.ReferenceId(flag.Arg(2)))
    d.CheckError("Failed to get length of set", e)

    fmt.Println(length)
}

func (d *CRDBGSetCommandListener) DoContains(client *crdb.Client) {
    d.CheckNArg(4, "contains <ReferenceId> <OBJECT_DATA>")

    result, e := client.GSetClient.Contains(crdb.ReferenceId(flag.Arg(2)), []byte(flag.Arg(3)))
    d.CheckError("Failed to check contains of set", e)

    fmt.Println(result)
}

func (d *CRDBGSetCommandListener) DoEquals(client *crdb.Client) {
    d.CheckNArg(4, "equals <ReferenceId> <ReferenceId>")

    result, e := client.GSetClient.Equals(crdb.ReferenceId(flag.Arg(2)), crdb.ReferenceId(flag.Arg(3)))
    d.CheckError("Failed to check set equality", e)

    fmt.Println(result)
}

func (d *CRDBGSetCommandListener) DoMerge(client *crdb.Client) {
    d.CheckNArg(4, "merge <ReferenceId> <ReferenceId>")

    e := client.GSetClient.Merge(crdb.ReferenceId(flag.Arg(2)), crdb.ReferenceId(flag.Arg(4)))
    d.CheckError("Failed to do merge operation on set", e)
}

func (d *CRDBGSetCommandListener) DoClone(client *crdb.Client) {
    d.CheckNArg(3, "clone <ReferenceId>")

    result, e := client.GSetClient.Clone(crdb.ReferenceId(flag.Arg(2)))
    d.CheckError("Failed to do clone operation on set", e)

    fmt.Println("ResourceId:", result)
}

func (d *CRDBGSetCommandListener) Execute(client *crdb.Client) {
    usage := "<list|insert|length|contains|equals|merge|clone>"

    if flag.NArg() < 2 {
        d.ShowUsage(usage)
        os.Exit(1)
    }

    switch flag.Arg(1) {
    case "list": d.DoList(client)
    case "insert": d.DoInsert(client)
    case "length": d.DoLength(client)
    case "contains": d.DoContains(client)
    case "equals": d.DoEquals(client)
    case "merge": d.DoMerge(client)
    case "clone": d.DoClone(client)
    default:
        d.ShowUsage(usage)
        os.Exit(1)
    }
}

