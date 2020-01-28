package ziptrans

import (
	"archive/zip"
	"encoding/binary"

	"go.polydawn.net/rio/fs"
)

// MetadataToZipHdr mutates zip.FileHeader fields to match the given fmeta.
func MetadataToZipHdr(fmeta *fs.Metadata, hdr *zip.FileHeader) {
	hdr.Name = fmeta.Name.String()
	if fmeta.Type == fs.Type_Dir {
		hdr.Name += "/"
	}
	hdr.UncompressedSize64 = uint64(fmeta.Size)
	hdr.Extra = append(zipUnix2ExtraHeader(fmeta), zipUnix3ExtraHeader(fmeta)...)
	hdr.SetMode(fmeta.Mode())
	hdr.SetModTime(fmeta.Mtime)
}

// compose a unix2 (0x7855) header for the zip file.
func zipUnix2ExtraHeader(fmeta *fs.Metadata) []byte {
	// Do not include the older unix2 header if the uid/gid don't fit in that representation.
	if uint32(uint16(fmeta.Uid)) != fmeta.Uid || uint32(uint16(fmeta.Gid)) != fmeta.Gid {
		return []byte{}
	}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint16(buf[0:2], 0x7855)
	binary.LittleEndian.PutUint16(buf[2:4], 4) // total data size for the block
	binary.LittleEndian.PutUint16(buf[4:6], uint16(fmeta.Uid))
	binary.LittleEndian.PutUint16(buf[6:8], uint16(fmeta.Gid))
	return buf
}

// compose a unix3 (0x7875) header for the zip file.
func zipUnix3ExtraHeader(fmeta *fs.Metadata) []byte {
	buf := make([]byte, 15)
	binary.LittleEndian.PutUint16(buf[0:2], 0x7875)
	binary.LittleEndian.PutUint16(buf[2:4], 11) // total data size for the block
	buf[4] = 1                                  // Version
	buf[5] = 5                                  // UIDSize
	binary.LittleEndian.PutUint32(buf[6:10], fmeta.Uid)
	buf[10] = 5 // GIDSize
	binary.LittleEndian.PutUint32(buf[11:15], fmeta.Gid)
	return buf
}
