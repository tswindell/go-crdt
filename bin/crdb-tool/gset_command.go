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

type CRDBGSetCommandListener struct {}

func (d *CRDBGSetCommandListener) RespondTo(cmd string) bool {
    return cmd == "crdt:gset" || cmd == "gset"
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
        fmt.Println(string(item))
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
    default:
        d.ShowUsage(usage)
        os.Exit(1)
    }
}

