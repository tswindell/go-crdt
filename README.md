# go-crdt

## Introduction

TBD

## Building

Firstly build the database daemon process:

```
  $ go get github.com/tswindell/go-crdt/bin/crdbd
```

Then we build the command line tool:

```
  $ go get github.com/tswindell/go-crdt/bin/crdb-tool
```

## Running

In one terminal run:
```
  $ crdbd
```

In another terminal:
```
  $ crdb-tool types
```

Which should output the following:

```
CRDT Supported Types:
  crdt:gset - true
  crdt:2pset - true
```

## CLI Tool Examples

Creating a new resource:
```
  $ crdb-tool create crdt:gset tmpfs aes-256-cbc
```

Attaching to resource, using *Id* and *Key* returned from *create*:
```
  $ crdb-tool attach <ResourceId> <ResourceKey>
```

### Manipulating GSet Resource
Insert a new object into the GSet, using *ReferenceId* from *attach*:
```
 $ crdb-tool crdt:gset insert <ReferenceId> <OBJECT_DATA>
```

List elements in a GSet.
```
 $ crdb-tool crdt:gset list <ReferenceId>
```

GSet Sub-Commands:
```
  * list <ReferenceId>
  * insert <ReferenceId> <OBJECT_DATA>
  * length <ReferenceId>
  * contains <ReferenceId> <OBJECT_DATA>
  * equals <ReferenceId> <OtherReferenceId>
  * merge <ReferenceId> <OtherReferenceId>
  * clone <ReferenceId>
```
