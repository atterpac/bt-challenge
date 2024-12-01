package packer

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Block struct {
	ID       int32          // Unique ID of the block
	Files    []FileMetadata // Files contained in the block
	Size     int64          // Current size of the block
	Checksum []byte         // SHA-256 checksum of the block
	Writer   io.Writer      // Writer for block content
}

// addFileToBlock adds a file to a block with corresponding metadata
func (p defaultPacker) addFileToBlock(block *Block, file *FileInfo) error {
	// Open file path
	f, err := os.Open(file.Path)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	// Calculate checksum
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("error calculating checksum for file: %w", err)
	}

	// Reset file pointer
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("error resetting file pointer in file: %w", err)
	}

	// Create metadata
	metaData := &FileMetadata{
		Path:     file.Path,
		Size:     file.Size,
		ModTime:  file.ModTime,
		Mode:     file.Mode,
		BlockID:  block.ID,
		Offset:   block.Size,
		Checksum: h.Sum(nil),
	}

	// Update block
	block.Files = append(block.Files, *metaData)
	block.Size += file.Size

	return nil
}

func (p defaultPacker) writeBlock(block *Block, outputDir string, blockNum int32) error {
	// Create block file
	blockPath := filepath.Join(outputDir, fmt.Sprintf("block-%d.beam", blockNum))
	f, err := os.Create(blockPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write block data
	h := sha256.New()
	w := io.MultiWriter(f, h)

	// Write block ID
	if err := binary.Write(w, binary.LittleEndian, int32(blockNum)); err != nil {
		return err
	}

	// Write number of files in block
	if err := binary.Write(w, binary.LittleEndian, int32(len(block.Files))); err != nil {
		return err
	}

	// Write metadata for each file
	for _, metadata := range block.Files {
		if err := p.writeMetadata(w, &metadata); err != nil {
			return err
		}
	}

	// Write file contents
	for _, metadata := range block.Files {
		f, err := os.Open(metadata.Path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", metadata.Path, err)
		}
		// Copy file contents to block
		if _, err := io.Copy(w, f); err != nil {
			return fmt.Errorf("failed to write file %s: %w", metadata.Path, err)
		}

		f.Close()
	}

	// Write block checksum
	blockChecksum := h.Sum(nil)
	if _, err := w.Write(blockChecksum); err != nil {
		return fmt.Errorf("failed to write block checksum: %w", err)
	}

	return nil
}

func (p *defaultPacker) extractFile(r io.Reader, outputDir string, metadata *FileMetadata) error {
	// Create output file
	outputPath := filepath.Join(outputDir, metadata.Path)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("error creating directory for file: %w", err)
	}

	// Open output file
	f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, os.FileMode(metadata.Mode))
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	w := io.MultiWriter(f, h)

	// Copy file contents
	if _, err := io.CopyN(w, r, metadata.Size); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	// Verify checksum
	if !bytes.Equal(h.Sum(nil), metadata.Checksum) {
		return fmt.Errorf("checksum mismatch for file: %s", metadata.Path)
	}

	// Set file modification time
	if err := os.Chtimes(outputPath, metadata.ModTime, metadata.ModTime); err != nil {
		return fmt.Errorf("error setting file modification time: %w", err)
	}

	return nil
}
