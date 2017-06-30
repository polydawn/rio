developer readme
================

(This document is for someone hacking on the `rio` code itself.  Go up one dir for user-facing documentation!)



big picture
-----------

`rio` handles data in two catagorically different ways:

- Sets of files, unpacked on a real filesystem on your host
- Packed "wares", which represent sets of files and all their metadata, but are organized in (furious handwaving and abstractions) some other way.

Packed "wares" are content-addressable: `WareID`s are hashes.

### universal IDs

... don't exist.  It's not a thing.  We give up.

Git hashes are content-addressable, and IPFS hashes are content-addressable, and they're not the same, and *that's okay*.

`rio.WareID` is a two-part object: it expresses the pack format, and the hash, *which is only defined within the namespace of that pack format*.

Example: `"git:asdf"` and `"ipfs:qwer"` are both valid WareIDs, and may even refer to the same files and metadata when unpacked.

*There is also no single hash which defines the WareID of an unpacked filesystem*.  This is just the other side of the same coin.
You can only compute a hash of a fileset by picking which packing format you want to use.



code layout
-----------

- `go.polydawn.net/rio` -- main package.  Interface definitions.  Other projects using `rio` as a library should import this package -- and few others.
- `go.polydawn.net/rio/fs` -- types for paths, and an abstract filesystem.  All `rio` code uses this when describing files and filesystem metadata.
  - `go.polydawn.net/rio/fs/osfs` -- a concrete implementation of the `fs` interfaces, implemented with a regular filesystem.
- `go.polydawn.net/rio/fsOp` -- operations on filesystems.  Distinct from `fs` because `fsOp` is more intention-oriented; `fs` is a fairly direct proxy to syscalls, and much less friendly.
- `go.polydawn.net/rio/warehouse/*` -- implementations of storage warehouses.  Local filesystem, S3, GCS, IPFS... each get their own package under here.
- `go.polydawn.net/transmat/*` -- implementations of filesystem packing formats.  E.g. `tar`.
  - REVIEW: so is this name a bug and the whole package should be `s/transmat/packing/`??  Probably
- `go.polydawn.net/transmat/mixins/fshash` -- helper functions for accumulating a hash for a fileset.  Used in some of the transmat implementations.
- `go.polydawn.net/lib/*` -- grabbag library functions; these are things that *probably* make sense even more broadly than rio, but are vendored here for simplicity's sake.

Overall, seen from the outside (as a consumer of `rio`-as-a-library):

- You want to handle Filesets.
  - You can handle them using the `fs` types for specifying paths
  - and you *may* use the `fs` types if necessary for inspecting or manipulating metadata...
    - but usually you want to use a transmat to handle filesets for you!
- Plug together a `packing` and a `warehouse` sytem to get a transmat.
  - for example `tar`+`s3`, or `tar`+`fs`: both construct valid transmats.
- And from the perspective of library consumers: that's it, no more details should need to leak.

### the fs vs fsOp split

Where 'fs' ends and 'fsOp' begins can be difficult to define, so here are some examples.

1. `Chown()` belongs in `fs`, because it's basically proxying a syscall.
2. `PlaceFile()` belongs in `fsOp` because it composes multiple syscalls...
3. More importantly, `PlaceFile()` belongs in `fsOp` because it implements some application-level sandboxing logic!  It has *opinions*.
