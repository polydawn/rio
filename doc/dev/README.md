developer readme
================

(This document is for someone hacking on the `rio` code itself.  Go up one dir for user-facing documentation!)


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

Overall, seen from the outside (as a consumer of `rio`-as-a-library):

- You want to handle Filesets.
  - You can handle them using the `fs` types for specifying paths
  - and you *may* use the `fs` types if necessary for inspecting or manipulating metadata...
    - but usually you want to use a transmat to handle filesets for you!
- Plug together a `packing` and a `warehouse` sytem to get a transmat.
  - for example `tar`+`s3`, or `tar`+`fs`: both construct valid transmats.
- And from the perspective of library consumers: that's it, no more details should need to leak.
