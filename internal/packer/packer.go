package packer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type defaultPacker struct {
	opts      PackerOptions
	validator *Validator
}

func NewPacker(opts PackerOptions) Packer {
	return defaultPacker{
		opts:      opts,
		validator: NewValidator(opts.BufferSize),
	}
}

func (p defaultPacker) Pack(inputDir string, outputDir string) error {
	// Walk files in inputDir
	var files []string
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking input directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found in input directory")
	}

	// Create outputDir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	fileInfos, err := p.collectFileInfo(files)
	if err != nil {
		return fmt.Errorf("error collecting file info: %w", err)
	}

	// Sort files by size
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].Size > fileInfos[j].Size
	})

	return p.packFiles(fileInfos, outputDir)
}

func (p defaultPacker) Unpack(inputDir string, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	info, err := os.Stat(inputDir)
	if err != nil {
		return fmt.Errorf("failed to get input directory info: %w", err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(inputDir)
		if err != nil {
			return fmt.Errorf("failed to read input directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".beam" {
				blockPath := filepath.Join(inputDir, entry.Name())
				if err := p.unpackBlock(blockPath, outputDir); err != nil {
					return fmt.Errorf("error unpacking block %s: %w", entry.Name(), err)
				}
			}
		}
		return nil
	}

	return p.unpackBlock(inputDir, outputDir)
}

func (p defaultPacker) Verify(inputDir string) error {
	info, err := os.Stat(inputDir)
	if err != nil {
		return fmt.Errorf("failed to get input directory info: %w", err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(inputDir)
		if err != nil {
			return fmt.Errorf("failed to read input directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".beam" {
				blockPath := filepath.Join(inputDir, entry.Name())
				if err := p.validator.ValidateBlock(blockPath); err != nil {
					return fmt.Errorf("error verifying block integrity: %w", err)
				}
			}
		}
		return nil
	}
	return p.validator.ValidateBlock(inputDir)
}

// collectFileInfo collects file info for all files in the input directory
func (p defaultPacker) collectFileInfo(files []string) ([]FileInfo, error) {
	var fileInfo []FileInfo

	for _, path := range files {

		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("error getting file info: %w", err)
		}

		if info.Size() > BlockSize {
			fmt.Printf("Skipping file %s, size exceeds block size\n", path)
			continue
		}

		if !info.IsDir() {
			fileInfo = append(fileInfo, FileInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime(),
				Mode:    uint32(info.Mode()),
				IsDir:   false,
			})
		}
	}
	return fileInfo, nil
}

func (p defaultPacker) packFiles(files []FileInfo, outputDir string) error {
	blockNum := int32(1)
	currentBlock := &Block{
		ID:    blockNum,
		Files: make([]FileMetadata, 0),
		Size:  0,
	}

	var currentSize int64

	for i, file := range files {
		// If file doesnt fit in currnt block, write current block and start new one
		if currentSize+file.Size > BlockSize {
			if err := p.writeBlock(currentBlock, outputDir, blockNum); err != nil {
				return fmt.Errorf("error writing block: %w", err)
			}
			blockNum++
			currentSize = 0
			currentBlock = &Block{
				ID:    blockNum,
				Files: make([]FileMetadata, 0),
				Size:  0,
			}
		}

		if err := p.addFileToBlock(currentBlock, &file); err != nil {
			return err
		}
		currentSize += file.Size

		if i == len(files)-1 {
			if err := p.writeBlock(currentBlock, outputDir, blockNum); err != nil {
				return fmt.Errorf("error writing block: %w", err)
			}
		}
	}
	return nil
}
