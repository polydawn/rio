package ziptrans

import (
	"archive/zip"
	"encoding/binary"

	. "github.com/warpfork/go-errcat"
	"go.polydawn.net/go-timeless-api/rio"
	"go.polydawn.net/rio/fs"
	"go.polydawn.net/rio/fs/osfs"
)

// MetadataToZipHdr mutates zip.FileHeader fields to match the given fmeta.
func MetadataToZipHdr(fmeta *fs.Metadata, hdr *zip.FileHeader) {
	hdr.Name = fmeta.Name.String()
	if fmeta.Type == fs.Type_Dir {
		hdr.Name += "/"
	}

	// compress the data.
	hdr.Method = zip.Deflate

	hdr.UncompressedSize64 = uint64(fmeta.Size)
	hdr.Extra = append(zipUnix2ExtraHeader(fmeta), zipUnix3ExtraHeader(fmeta)...)
	hdr.SetMode(osfs.ModeToOs(fmeta))
	hdr.SetModTime(fmeta.Mtime)
}

type zipExtraHeaderID uint16

const (
	zipExtraUnix2 = zipExtraHeaderID(0x7855)
	zipExtraUnix3 = zipExtraHeaderID(0x7875)
)

type zipExtraHeader struct {
	length uint16
	data   []byte
}

// compose a unix2 (0x7855) header for the zip file.
func zipUnix2ExtraHeader(fmeta *fs.Metadata) []byte {
	// Do not include the older unix2 header if the uid/gid don't fit in that representation.
	if uint32(uint16(fmeta.Uid)) != fmeta.Uid || uint32(uint16(fmeta.Gid)) != fmeta.Gid {
		return []byte{}
	}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint16(buf[0:2], uint16(zipExtraUnix2))
	binary.LittleEndian.PutUint16(buf[2:4], 4) // total data size for the block
	binary.LittleEndian.PutUint16(buf[4:6], uint16(fmeta.Uid))
	binary.LittleEndian.PutUint16(buf[6:8], uint16(fmeta.Gid))
	return buf
}

func parseUnix2Header(hdr []byte) (uint32, uint32, error) {
	if len(hdr) < 8 {
		return 0, 0, Errorf(rio.ErrWareCorrupt, "Corrupt zip File Header. Invalid Unix2 Extra Header")
	}
	uid := binary.LittleEndian.Uint16(hdr[4:6])
	gid := binary.LittleEndian.Uint16(hdr[6:8])
	return uint32(uid), uint32(gid), nil
}

// compose a unix3 (0x7875) header for the zip file.
func zipUnix3ExtraHeader(fmeta *fs.Metadata) []byte {
	buf := make([]byte, 15)
	binary.LittleEndian.PutUint16(buf[0:2], uint16(zipExtraUnix3))
	binary.LittleEndian.PutUint16(buf[2:4], 11) // total data size for the block
	buf[4] = 1                                  // Version
	buf[5] = 4                                  // UIDSize
	binary.LittleEndian.PutUint32(buf[6:10], fmeta.Uid)
	buf[10] = 4 // GIDSize
	binary.LittleEndian.PutUint32(buf[11:15], fmeta.Gid)
	return buf
}

func parseUnix3Header(hdr []byte) (uint32, uint32, error) {
	if len(hdr) < 7 {
		return 0, 0, Errorf(rio.ErrWareCorrupt, "Corrupt zip File Header. Too Short Unix3 Extra Header")
	}
	uidSize := hdr[1]
	uid := uint32(0)
	if uidSize == 2 {
		uid = uint32(binary.LittleEndian.Uint16(hdr[2:4]))
	} else if uidSize == 4 {
		uid = uint32(binary.LittleEndian.Uint32(hdr[2:6]))
	} else {
		return 0, 0, Errorf(rio.ErrWareCorrupt, "Corrupt zip File Header. Invalid Unix3 Extra Header UID Length: %d", uidSize)
	}

	gidSize := hdr[2+uidSize]
	gid := uint32(0)
	if gidSize == 2 {
		gid = uint32(binary.LittleEndian.Uint16(hdr[3+uidSize : 3+uidSize+2]))
	} else if gidSize == 4 {
		gid = uint32(binary.LittleEndian.Uint32(hdr[3+uidSize : 3+uidSize+4]))
	} else {
		return 0, 0, Errorf(rio.ErrWareCorrupt, "Corrupt zip File Header. Invalid Unix3 Extra Header GID length")
	}

	return uid, gid, nil
}

func parseZipExtraHeader(hdr *zip.FileHeader) (map[zipExtraHeaderID]zipExtraHeader, error) {
	m := make(map[zipExtraHeaderID]zipExtraHeader)
	i := 0
	for i < len(hdr.Extra)-3 {
		id := binary.LittleEndian.Uint16(hdr.Extra[i : i+2])
		l := binary.LittleEndian.Uint16(hdr.Extra[i+2 : i+4])
		if len(hdr.Extra) < i+int(l)+4 {
			return m, Errorf(rio.ErrWareCorrupt, "Corrupt zip File Header: %s", hdr.Name)
		}
		m[zipExtraHeaderID(id)] = zipExtraHeader{
			length: l,
			data:   hdr.Extra[i+4 : i+4+int(l)],
		}
		i = i + 4 + int(l)
	}
	return m, nil
}

// Look for UID/GID bits stored in extra headers.
func zipFileOwnership(hdr *zip.FileHeader) (uint32, uint32, error) {
	hdrs, err := parseZipExtraHeader(hdr)
	if err != nil {
		return 0, 0, err
	}
	if unix3, ok := hdrs[zipExtraUnix3]; ok {
		return parseUnix3Header(unix3.data)
	}
	if unix2, ok := hdrs[zipExtraUnix2]; ok {
		return parseUnix2Header(unix2.data)
	}

	// TODO: better default.
	return 1000, 1000, nil
}

// ZipHdrToMetadata mutates fs.Metadata fields to match the given zip header.
func ZipHdrToMetadata(hdr *zip.FileHeader, fmeta *fs.Metadata) (skipMe error, haltMe error) {
	finfo := hdr.FileInfo()

	fmeta.Name = fs.MustRelPath(hdr.Name) // FIXME should not use the 'must' path
	fmeta.Perms = osfs.OsToPerms(finfo.Mode())
	fmeta.Type = osfs.OsToType(finfo.Mode())
	if fmeta.Type == fs.Type_Invalid {
		return nil, Errorf(rio.ErrWareCorrupt, "corrupt zip: %q is not of a known file type", hdr.Name)
	}
	uid, gid, err := zipFileOwnership(hdr)
	if err != nil {
		return nil, Errorf(rio.ErrWareCorrupt, "corrupt zip: %q does not specify owner: %v", hdr.Name, err)
	}
	fmeta.Uid = uid
	fmeta.Gid = gid
	fmeta.Size = finfo.Size()
	// fmeta.Linkname isn't retrieved from the FileHeader, but rather from reading the file contents.
	// TODO: devices (fmeta.Devmajor / Devminor)
	fmeta.Mtime = hdr.Modified
	//TODO: xattrs
	return nil, nil
}
