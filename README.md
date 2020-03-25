rio
===

**R**epeatable **I**/**O**.

Rio is tooling for repeatable, reproducible, filesystem I/O and transport.
It's both a command line tool, and usable as a series of libraries.

Rio is part of the [Timeless Stack](https://repeatr.io) --
it handles packing, unpacking, syncing, copying, and hashing of Filesets.

You may also want to check out [Repeatr](https://github.com/polydawn/repeatr),
which is the container executor in the Timeless Stack (it uses Rio under the hood
to provide snapshotting and decentralized sync of container filesystem images);
and [Reach](https://github.com/polydawn/reach), which provides pipelining tooling
for managing both Rio and Repeatr with less copy-pasting of hashes.


What?
-----

### example of packing

Okay, by example:

```
rio pack tar /tmp/something/ --target=file://output.tgz
```

That just created a tar pack of the files you aimed it at.

It also gave you a hash on stdout.  That's a
[content-addressable](https://en.wikipedia.org/wiki/Content-addressable_storage)
ID of the pack you just created.

### example of unpacking

You can use that ID later to unpack things:

```
rio unpack tar:879UrF8j7E[...]udF57KpF8 \
   /tmp/unpackhere/ --source=file://output.tgz
```

It's long.  That's because it's a [cryptographic hash](https://en.wikipedia.org/wiki/Cryptographic_hash_function).

Why is this neat?

You can put that packed data on any server, anywhere, and fetch it again by hash.
The hash makes it immutable, and reproducible, even if you don't control the storage.

### example of packing to remote storage and using content-addressing

```
rio pack tar /tmp/something/ --target=ca+https://mybucket.s3.amazonaws.com/assets/
```

Same drill except... we just:

- uploaded to a cloud storage host
- without a specific name -- the hash is used to organize storage (note the "ca+" segment of the url).
- and still got the same hash.

### example of unpacking from remote storage while using content-addressing

Now, hand that hash to someone else:

```
rio unpack tar:879UrF8j7E[...]udF57KpF8 \
   /tmp/unpackhere/ --source=ca+https:///mybucket.s3.amazonaws.com/assets/
```

They get your files.  Boom.  Huge amounts of data.  Just one handle to copypaste: that hash.


Building
--------

Rio is built in Golang and uses git submodules to track libraries by hash.

To build Rio, first get the submodules, then set up GOPATH, then use go:

```
git submodule update --init
GOPATH=$PWD/.gopath go install ./...
```

You may find the [`gof`](https://github.com/warpfork/gof) script makes this more convenient.
