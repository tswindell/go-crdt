package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/tswindell/go-crdt/db"
)

type TwoPhaseSetCommandListener struct {}

func (d *TwoPhaseSetCommandListener) RespondTo(cmd string) bool {
    return cmd == "crdt:2pset" || cmd == "2pset"
}

func (d *TwoPhaseSetCommandListener) ShowUsage(usage string) {
    fmt.Fprintf(os.Stderr, "Usage: crdb-tool crdt:2pset %s\n", usage)
}

func (d *TwoPhaseSetCommandListener) CheckNArg(count int, usage string) {
    if flag.NArg() < count {
        d.ShowUsage(usage)
        os.Exit(1)
    }
}

func (d *TwoPhaseSetCommandListener) CheckError(m string, e error) {
    if e != nil {
        fmt.Fprintf(os.Stderr, "Error: %s: %v\n", m, e)
        os.Exit(1)
    }
}

func (d *TwoPhaseSetCommandListener) DoList(client *crdb.Client) {
    d.CheckNArg(3, "list <ReferenceId>")

    ch, e := client.TwoPhaseSetClient.List(crdb.ReferenceId(flag.Arg(2)))
    d.CheckError("Failed to list set", e)

    for item := range ch {
        fmt.Println(string(item))
    }
}

func (d *TwoPhaseSetCommandListener) DoInsert(client *crdb.Client) {
    d.CheckNArg(4, "insert <ReferenceId> <OBJECT_DATA>")

    e := client.TwoPhaseSetClient.Insert(crdb.ReferenceId(flag.Arg(2)), []byte(flag.Arg(3)))
    d.CheckError("Failed to insert item", e)
}

func (d *TwoPhaseSetCommandListener) DoRemove(client *crdb.Client) {
    d.CheckNArg(4, "remove <ReferenceId> <OBJECT_DATA>")

    e := client.TwoPhaseSetClient.Remove(crdb.ReferenceId(flag.Arg(2)), []byte(flag.Arg(3)))
    d.CheckError("Failed to remove item", e)
}

func (d *TwoPhaseSetCommandListener) DoLength(client *crdb.Client) {
    d.CheckNArg(3, "length <ReferenceId>")

    length, e := client.TwoPhaseSetClient.Length(crdb.ReferenceId(flag.Arg(2)))
    d.CheckError("Failed to get length of set", e)

    fmt.Println(length)
}

func (d *TwoPhaseSetCommandListener) DoContains(client *crdb.Client) {
    d.CheckNArg(4, "contains <ReferenceId> <OBJECT_DATA>")

    result, e := client.TwoPhaseSetClient.Contains(crdb.ReferenceId(flag.Arg(2)), []byte(flag.Arg(3)))
    d.CheckError("Failed to check contains of set", e)

    fmt.Println(result)
}

func (d *TwoPhaseSetCommandListener) DoEquals(client *crdb.Client) {
    d.CheckNArg(4, "equals <ReferenceId> <ReferenceId>")

    result, e := client.TwoPhaseSetClient.Equals(crdb.ReferenceId(flag.Arg(2)), crdb.ReferenceId(flag.Arg(3)))
    d.CheckError("Failed to check set equality", e)

    fmt.Println(result)
}

func (d *TwoPhaseSetCommandListener) DoMerge(client *crdb.Client) {
    d.CheckNArg(4, "merge <ReferenceId> <ReferenceId>")

    e := client.TwoPhaseSetClient.Merge(crdb.ReferenceId(flag.Arg(2)), crdb.ReferenceId(flag.Arg(4)))
    d.CheckError("Failed to do merge operation on set", e)
}

func (d *TwoPhaseSetCommandListener) DoClone(client *crdb.Client) {
    d.CheckNArg(3, "clone <ReferenceId>")

    resourceId, resourceKey, e := client.TwoPhaseSetClient.Clone(crdb.ReferenceId(flag.Arg(2)))
    d.CheckError("Failed to do clone operation on set", e)

    fmt.Println("ResourceId:", resourceId)
    fmt.Println("ResourceKey:", resourceKey)
}

func (d *TwoPhaseSetCommandListener) Execute(client *crdb.Client) {
    usage := "<list|insert|length|contains|equals|merge|clone>"

    if flag.NArg() < 2 {
        d.ShowUsage(usage)
        os.Exit(1)
    }

    switch flag.Arg(1) {
    case "list": d.DoList(client)
    case "insert": d.DoInsert(client)
    case "remove": d.DoRemove(client)
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

