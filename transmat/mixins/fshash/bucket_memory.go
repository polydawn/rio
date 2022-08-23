package fshash

import (
	"fmt"
	"sort"
	"strings"

	"github.com/polydawn/rio/fs"
	"github.com/polydawn/rio/lib/treewalk"
)

var _ Bucket = &MemoryBucket{}

type MemoryBucket struct {
	// my kingdom for a red-black tree or other sane sorted map implementation
	names   []string
	records map[string]Record
}

func (b *MemoryBucket) AddRecord(metadata fs.Metadata, contentHash []byte) {
	if b.records == nil {
		b.records = map[string]Record{}
	}

	name := metadata.Name.String()
	if metadata.Type == fs.Type_Dir {
		name += "/"
	}
	b.names = append(b.names, name)
	b.records[name] = Record{name, metadata, contentHash}
}

func (b *MemoryBucket) HasRecord(metadata fs.Metadata) bool {
	name := metadata.Name.String()
	if metadata.Type == fs.Type_Dir {
		name += "/"
	}
	_, hasRecord := b.records[name]
	return hasRecord
}

func (b *MemoryBucket) UpdateRecord(metadata fs.Metadata, contentHash []byte) {
	name := metadata.Name.String()
	if metadata.Type == fs.Type_Dir {
		name += "/"
	}
	b.records[name] = Record{name, metadata, contentHash}
}

func (b *MemoryBucket) recordList() []Record {
	var records []Record
	for _, n := range b.names {
		records = append(records, b.records[n])
	}
	return records
}

/*
	Get a `treewalk.Node` that starts at the root of the bucket.
	The walk will be in deterministic, sorted order (and thus is appropriate
	for hashing).

	This applies some "finalization" operations before starting the walk:
	  - All records will be sorted.
	  - As a sanity check, if records exist, the first one must be ".".

	This is only safe for non-concurrent use and depth-first traversal.
	If the data structure is changed, or (sub)iterators used out of order,
	behavior is undefined.
*/
func (b *MemoryBucket) Iterator() RecordIterator {
	sort.Sort(memoryBucketByFilename(*b))
	if len(b.records) > 0 {
		firstPath := b.Root().Metadata.Name
		if firstPath != (fs.RelPath{}) {
			panic(ErrInvalidFilesystem{fmt.Sprintf("missing root (first entry: %q)", firstPath)})
		}
	}
	var that int
	return &memoryBucketIterator{b.recordList(), 0, &that}
}

func (b *MemoryBucket) Root() Record {
	return b.records[b.names[0]]
}

func (b *MemoryBucket) Length() int {
	return len(b.records)
}

type memoryBucketIterator struct {
	lines []Record
	this  int  // pretending a linear structure is a tree is weird.
	that  *int // this is the last child walked.
}

func (i *memoryBucketIterator) NextChild() treewalk.Node {
	// Since we sorted before starting iteration, all child nodes are contiguous and follow their parent.
	// Each treewalk node keeps its own record's index (implicitly, this is forming a stack),
	// and they all share the same value for last index walked, so when a child has been fully iterated over,
	// the next call on the parent will start looking right after all the child's children.
	next := *i.that + 1
	if next >= len(i.lines) {
		return nil
	}
	nextName := i.lines[next].Name
	thisName := i.lines[i.this].Name
	// is the next one still a child?
	if strings.HasPrefix(nextName, thisName) {
		// check for repeated names
		if i.lines[*i.that].Name == nextName {
			panic(ErrInvalidFilesystem{fmt.Sprintf("repeated path: %q", nextName)})
		}
		// check for missing trees
		if strings.ContainsRune(nextName[len(thisName):len(nextName)-1], '/') {
			panic(ErrInvalidFilesystem{fmt.Sprintf("missing tree: %q followed %q", nextName, thisName)})
		}
		// step forward
		*i.that = next
		return &memoryBucketIterator{i.lines, *i.that, i.that}
	}
	return nil
}

func (i memoryBucketIterator) Record() Record {
	return i.lines[i.this]
}

type memoryBucketByFilename MemoryBucket

func (a memoryBucketByFilename) Len() int           { return len(a.records) }
func (a memoryBucketByFilename) Swap(i, j int)      { a.names[i], a.names[j] = a.names[j], a.names[i] }
func (a memoryBucketByFilename) Less(i, j int) bool { return a.names[i] < a.names[j] }
