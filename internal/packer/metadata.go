package packer

import (
	"crypto/sha256"
	"encoding/binary"
	"io"
	"time"
)

type FileMetadata struct {
	Path     string    // Original path
	Size     int64     // File size in bytes
	ModTime  time.Time // Last modification time
	Checksum []byte    // SHA-256 checksum of the file
	Offset   int64     // Offset within the block
	BlockID  int32     // ID of the block containing the file
	Mode     uint32    // File permissions
}

// FileInfo represents information about a file that is being processed
type FileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	Mode    uint32
	IsDir   bool
}

func (p *defaultPacker) writeMetadata(w io.Writer, metadata *FileMetadata) error {
	pathBytes := []byte(metadata.Path)
	if err := binary.Write(w, binary.LittleEndian, int32(len(pathBytes))); err != nil {
		return err
	}

	if _, err := w.Write(pathBytes); err != nil {
		return err
	}

	// Write Size
	if err := binary.Write(w, binary.LittleEndian, metadata.Size); err != nil {
		return err
	}

	// Write ModTime
	if err := binary.Write(w, binary.LittleEndian, metadata.ModTime.Unix()); err != nil {
		return err
	}

	// Write Offset
	if err := binary.Write(w, binary.LittleEndian, metadata.Offset); err != nil {
		return err
	}

	// Write Mode
	if err := binary.Write(w, binary.LittleEndian, metadata.Mode); err != nil {
		return err
	}

	// Write Checksum
	if _, err := w.Write(metadata.Checksum); err != nil {
		return err
	}

	return nil
}

func (p *defaultPacker) readMetadata(r io.Reader) (*FileMetadata, error) {
	var pathLen int32
	if err := binary.Read(r, binary.LittleEndian, &pathLen); err != nil {
		return nil, err
	}

	// Get Path
	pathBytes := make([]byte, pathLen)
	if _, err := r.Read(pathBytes); err != nil {
		return nil, err
	}

	size := int64(0)
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err
	}

	var modTime int64
	if err := binary.Read(r, binary.LittleEndian, &modTime); err != nil {
		return nil, err
	}

	var offset int64
	if err := binary.Read(r, binary.LittleEndian, &offset); err != nil {
		return nil, err
	}

	var mode uint32
	if err := binary.Read(r, binary.LittleEndian, &mode); err != nil {
		return nil, err
	}

	checksum := make([]byte, sha256.Size)
	if _, err := r.Read(checksum); err != nil {
		return nil, err
	}

	return &FileMetadata{
		Path:     string(pathBytes),
		Size:     size,
		ModTime:  time.Unix(modTime, 0),
		Offset:   offset,
		Mode:     mode,
		Checksum: checksum,
	}, nil
}
