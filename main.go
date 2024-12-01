package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/atterpac/bt-takehome/internal/packer"
)

var (
	DIR         = "test-generator/dist/sample-files" // Directory to pack
	OUTPUT_DIR  = "output"                           // Where to output packed files
	UNPACK_DIR  = "output/unpack"                    // Where to unpack files
	BUFFER_SIZE = 32 * 1024                          // 32KB buffer size for validation and checksum
	BLOCK_SIZE  = 60 * 1024 * 1024                   // 60MB block size
)

func main() {
	// Check arguments for overriding directories
	checkArgs()
	// Generate test files if they don't exist
	generateTestFiles()

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
		BlockSize:       int64(BLOCK_SIZE),
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

func checkArgs() {
	if len(os.Args) > 1 {
		println("Overriding default directories...")
		DIR = os.Args[1]
	}
	if len(os.Args) > 2 {
		println("Overriding default output directory...")
		OUTPUT_DIR = os.Args[2]
	}
	if len(os.Args) > 3 {
		println("Overriding default unpack directory...")
		UNPACK_DIR = os.Args[3]
	}
}

func generateTestFiles() {
	// Check if test-generator/dist/sample-files exists
	if _, err := os.Stat(DIR); os.IsNotExist(err) {
		fmt.Printf("Error: Generated tests not found in %s\n", DIR)
		// Generate sample files
		fmt.Println("Generating sample files...")
		// cd ./test-generator && go run main.go sample-files.yml
		if err := os.Chdir("test-generator"); err != nil {
			fmt.Printf("Error changing directory to test-generator: %v\n", err)
			os.Exit(1)
		}
		err := exec.Command("go", "run", "main.go", "sample-files.yml").Run()
		if err != nil {
			fmt.Printf("Error generating sample files: %v\n", err)
		}
		os.Chdir("..")
	}
}
