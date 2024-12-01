# File Packing System

This project implements a file packing system that efficiently packs multiple files into 60MB blocks. It provides functionality for packing files, verifying their integrity, and unpacking them back to their original structure.

## Features

- Pack multiple files into 60MB blocks
- Skip files larger than 60MB
- Data integrity verification
- File restoration to original structure

## Quick Start

1. Run the program using the following command in the project directory:
```bash
go run main.go
```
The program will generate sample files if they dont exist 

Pack them into blocks of `block-N.beam` files in `./output`

Unpack the blocks from  `./output` to `./output/unpack`

Verify integrity of the unpacked files

2. Optional arguments
```bash
go run main.go <input_dir> <output_dir> <unpack_dir>
```

## Algorithm Overview

The file packing system uses the following algorithm:

1. **File Discovery**:
   - Scan input directory recursively
   - Filter out files larger than 60MB
   - Collect file metadata (size, path, etc.)

2. **Packing Strategy**:
   - Sort files by size (largest first)
   - Use a "First-Fit" approach to pack files into 60MB blocks
   - Maintain original file structure information

3. **Block Format**:
   - Header: Contains metadata about packed files
   - Body: Contains actual file data
   - Footer: Checksum for integrity verification

4. **Integrity Validation**:
   - SHA-256 checksums for individual files
   - Block-level integrity checking

## Block Format

Each 60MB block is structured as follows:

### Block Header (8 bytes)
- Block ID (4 bytes): Unique identifier for the block
- Number of Files (4 bytes): Count of files in this block

### File Metadata Section (Variable size)
For each file:
- Path Length (4 bytes): Length of the file path string
- Path (variable): Original file path
- Size (8 bytes): File size in bytes
- ModTime (8 bytes): Last modification time (Unix timestamp)
- Offset (8 bytes): File's offset within the data section
- Mode (4 bytes): File permissions and mode
- Checksum (32 bytes): SHA-256 hash of file contents

### File Data Section (Variable size)
- Concatenated file contents in the order specified by metadata
- Each file starts at its specified offset
- Total section size ≤ 60MB

### Block Footer (32 bytes)
- Block Checksum (32 bytes): SHA-256 hash of entire block

This format ensures:
- Efficient file lookup and extraction
- Data integrity verification
- Preservation of file metadata
- Easy block validation

## Design Considerations

- Memory efficient processing of large files
- Data integrity preservation
- Clean separation of concerns

## Potential Improvements

1. Compression support
2. Parallel processing for large datasets
3. More sophisticated packing algorithms or optional algorithms options 
4. Encryption support
5. Streaming support for large files
6. Deduplication of identical files
7. Account for Block header and footer size for more accurate packing
9. Methods to track block usage statistics (Average utilization, Average file count, etc.)

