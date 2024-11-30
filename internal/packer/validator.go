package packer

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type BlockIntegrityError struct {
	BlockID     int
	ExpectedSum []byte
	ActualSum   []byte
}

func (e *BlockIntegrityError) Error() string {
	return fmt.Sprintf("block %d checksum mismatch: expected %x, got %x", e.BlockID, e.ExpectedSum, e.ActualSum)
}

type FileIntegrityError struct {
	Path        string
	ExpectedSum []byte
	ActualSum   []byte
}

func (e *FileIntegrityError) Error() string {
	return fmt.Sprintf("file %s checksum mismatch: expected %x, got %x", e.Path, e.ExpectedSum, e.ActualSum)
}

type Validator struct {
	bufferSize int
}

func NewValidator(bufferSize int) *Validator {
	return &Validator{
		bufferSize: bufferSize,
	}
}

func (v *Validator) VerifyFileIntegrity(path string, expectedSum []byte) error {
	actualSum, err := v.CalculateFileChecksum(path)
	if err != nil {
		return fmt.Errorf("error calculating checksum: %w", err)
	}
	if !v.ChecksumsEqual(expectedSum, actualSum) {
		return &FileIntegrityError{
			Path:        path,
			ExpectedSum: expectedSum,
			ActualSum:   actualSum,
		}
	}
	return nil
}

func (v *Validator) CalculateFileChecksum(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	return v.CalculateReaderChecksum(f)
}

func (v *Validator) CalculateReaderChecksum(r io.Reader) ([]byte, error) {
	h := sha256.New()
	buf := make([]byte, v.bufferSize)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, err := h.Write(buf[:n]); err != nil {
				return nil, fmt.Errorf("error calculating checksum: %w", err)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading file: %w", err)
		}
	}
	return h.Sum(nil), nil
}

func (v *Validator) ValidateBlock(blockPath string) error {
	f, err := os.Open(blockPath)
	if err != nil {
		return fmt.Errorf("error opening block file: %w", err)
	}
	defer f.Close()

	var blockID int32
	if err := binary.Read(f, binary.LittleEndian, &blockID); err != nil {
		return fmt.Errorf("error reading block ID: %w", err)
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}

	var storedChecksum [32]byte
	if _, err := f.Seek(-32, io.SeekEnd); err != nil {
		return fmt.Errorf("error seeking to checksum: %w", err)
	}
	if _, err := io.ReadFull(f, storedChecksum[:]); err != nil {
		return fmt.Errorf("error reading stored checksum: %w", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("error seeking to start of block: %w", err)
	}

	h := sha256.New()
	if _, err := io.CopyN(h, f, fileInfo.Size()-32); err != nil {
		return fmt.Errorf("error calculating checksum: %w", err)
	}

	actualChecksum := h.Sum(nil)

	if !v.ChecksumsEqual(storedChecksum[:], actualChecksum) {
		return &BlockIntegrityError{
			BlockID:     int(blockID),
			ExpectedSum: storedChecksum[:],
			ActualSum:   actualChecksum,
		}
	}
	return nil
}

func (v *Validator) ChecksumsEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, av := range a {
		if av != b[i] {
			return false
		}
	}
	return true
}
