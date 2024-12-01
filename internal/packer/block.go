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

func (p defaultPacker) addFileToBlock(block *Block, file *FileInfo) error {
	f, err := os.Open(file.Path)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("error calculating checksum for file: %w", err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("error resetting file pointer in file: %w", err)
	}

	metaData := &FileMetadata{
		Path:     file.Path,
		Size:     file.Size,
		ModTime:  file.ModTime,
		Mode:     file.Mode,
		BlockID:  block.ID,
		Offset:   block.Size,
		Checksum: h.Sum(nil),
	}

	block.Files = append(block.Files, *metaData)
	block.Size += file.Size

	return nil
}

func (p defaultPacker) writeBlock(block *Block, outputDir string, blockNum int32) error {
	blockPath := filepath.Join(outputDir, fmt.Sprintf("block-%d.beam", blockNum))
	f, err := os.Create(blockPath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	w := io.MultiWriter(f, h)

	if err := binary.Write(w, binary.LittleEndian, int32(blockNum)); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, int32(len(block.Files))); err != nil {
		return err
	}

	for _, metadata := range block.Files {
		if err := p.writeMetadata(w, &metadata); err != nil {
			return err
		}
	}

	for _, metadata := range block.Files {
		f, err := os.Open(metadata.Path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", metadata.Path, err)
		}
		if _, err := io.Copy(w, f); err != nil {
			return fmt.Errorf("failed to write file %s: %w", metadata.Path, err)
		}

		f.Close()
	}

	blockChecksum := h.Sum(nil)
	if _, err := w.Write(blockChecksum); err != nil {
		return fmt.Errorf("failed to write block checksum: %w", err)
	}

	return nil
}

func (p *defaultPacker) UnpackBlock(blockPath string, outputDir string) error {
	// Verify block integrity
	if p.opts.VerifyIntegrity {
		if err := p.validator.ValidateBlock(blockPath); err != nil {
			return fmt.Errorf("error verifying block integrity: %w", err)
		}
	}

	f, err := os.Open(blockPath)
	if err != nil {
		return fmt.Errorf("error opening block file: %w", err)
	}
	defer f.Close()

	var blockID int32
	if err := binary.Read(f, binary.LittleEndian, &blockID); err != nil {
		return err
	}

	var numFiles int32
	if err := binary.Read(f, binary.LittleEndian, &numFiles); err != nil {
		return fmt.Errorf("error reading number of files in block: %w", err)
	}

	files := make([]FileMetadata, numFiles)
	for i := range files {
		metadata, err := p.readMetadata(f)
		if err != nil {
			return fmt.Errorf("error reading metadata for file %d: %w", i, err)
		}
		files[i] = *metadata
	}

	for _, metadata := range files {
		if err := p.extractFile(f, outputDir, &metadata); err != nil {
			return fmt.Errorf("error extracting file %s: %w", metadata.Path, err)
		}
	}

	return nil
}

func (p *defaultPacker) extractFile(r io.Reader, outputDir string, metadata *FileMetadata) error {
	outputPath := filepath.Join(outputDir, metadata.Path)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("error creating directory for file: %w", err)
	}

	f, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY, os.FileMode(metadata.Mode))
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	w := io.MultiWriter(f, h)

	if _, err := io.CopyN(w, r, metadata.Size); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	// Verify checksum
	if !bytes.Equal(h.Sum(nil), metadata.Checksum) {
		return fmt.Errorf("checksum mismatch for file: %s", metadata.Path)
	}

	if err := os.Chtimes(outputPath, metadata.ModTime, metadata.ModTime); err != nil {
		return fmt.Errorf("error setting file modification time: %w", err)
	}

	return nil
}
