package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/atterpac/bt-takehome/internal/packer"
)

const (
	DIR         = "test-generator/dist/sample-files" // Directory to pack
	OUTPUT_DIR  = "output"                           // Base output directory
	UNPACK_DIR  = "output/unpack"                    // Directory for unpacked files
	BUFFER_SIZE = 32 * 1024                          // 32KB buffer size
	BLOCK_SIZE  = 60 * 1024 * 1024                   // 60MB block size
)

func main() {
	// Create output directories
	for _, dir := range []string{OUTPUT_DIR, UNPACK_DIR} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Create packer instance with integrity verification
	p := packer.NewPacker(packer.PackerOptions{
		VerifyIntegrity: true,
		BufferSize:      BUFFER_SIZE,
		BlockSize:       BLOCK_SIZE,
	})

	// Calculate original size
	originalSize := calculateTotalSize(DIR)
	fmt.Printf("\n=== Packing Stats ===\n")
	fmt.Printf("Source Directory: %s\n", DIR)

	// Pack files
	fmt.Printf("\nPacking files into %s...\n", OUTPUT_DIR)
	packStart := time.Now()
	if err := p.Pack(DIR, OUTPUT_DIR); err != nil {
		fmt.Printf("Error packing files: %v\n", err)
		os.Exit(1)
	}

	packDuration := time.Since(packStart)

	// Print packing stats
	fmt.Printf("\nPack Time: %v\n", packDuration)
	fmt.Printf("Pack Speed: %.2f MB/s\n", calculateSpeed(originalSize, packDuration))

	// Unpack files
	fmt.Printf("\nUnpacking files to %s...\n", UNPACK_DIR)
	unpackStart := time.Now()
	if err := p.Unpack(OUTPUT_DIR, UNPACK_DIR); err != nil {
		fmt.Printf("Error unpacking files: %v\n", err)
		os.Exit(1)
	}
	unpackDuration := time.Since(unpackStart)

	// Calculate unpacked size for verification
	unpackedSize := calculateTotalSize(UNPACK_DIR)

	// Print unpacking stats
	fmt.Printf("Unpack Time: %v\n", unpackDuration)
	fmt.Printf("Unpack Speed: %.2f MB/s\n", calculateSpeed(unpackedSize, unpackDuration))

	// Verify integrity
	fmt.Println("\nVerifying file integrity...")
	verifyStart := time.Now()
	if err := p.Verify(OUTPUT_DIR); err != nil {
		fmt.Printf("Verification failed: %v\n", err)
		runtime.Goexit()
	}
	fmt.Printf("Verification Time: %v\n", time.Since(verifyStart))
	fmt.Println("All files verified successfully!")
}

// calculateTotalSize calculates the total size of all files in a directory
func calculateTotalSize(dir string) int64 {
	var totalSize int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	return totalSize
}

// calculateSpeed calculates the processing speed in MB/s
func calculateSpeed(totalBytes int64, duration time.Duration) float64 {
	return float64(totalBytes) / (1024 * 1024) / duration.Seconds()
}
