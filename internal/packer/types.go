package packer

import (
	"io"
	"time"
)

const (
	// BlockSize is the maxium size of a block in bytes (60MB)
	BlockSize = 60 * 1024 * 1024
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

type Block struct {
	ID       int32          // Unique ID of the block
	Files    []FileMetadata // Files contained in the block
	Size     int64          // Current size of the block
	Checksum []byte         // SHA-256 checksum of the block
	Writer   io.Writer      // Writer for block content
}

// Packer defines the interface for file packing operations
type Packer interface {
	// Pack takes an input directory and packs all files into blocks in the output directory
	Pack(inputDir string, outputDir string) error

	// Unpack extracts files from blocks in the input and writes them to the output directory
	Unpack(inputDir string, outputDir string) error

	// Verify checks the integrity of the packed files
	Verify(inputDir string) error
}

// FileInfo represents information about a file that is being processed
type FileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	Mode    uint32
	IsDir   bool
}

// BlockWriter handles writing files into a block
type BlockWriter interface {
	// WriteFile writes a file into the current block
	WriteFile(fileInfo FileInfo, reader io.Reader) (FileMetadata, error)

	// Close finalizes the block and writes any remaining data
	Close() error
}

// BlockReader Handles reading files from a block
type BlockReader interface {
	// ReadFile reads a file from the block and writes it to the output directory
	ReadFile(metadata FileMetadata, outputDir string) error

	// Close closes the block reaer
	Close() error
}

// PackerOptions configures the behavior of the packer
type PackerOptions struct {
	VerifyIntegrity bool // Verify the integrity of the files after packing
	BufferSize      int  // Size of the buffer used for reading and writing files
	// Concurrent      bool // Enable concurrent processing
	// UseCompression bool // Use compression for the block files

}
