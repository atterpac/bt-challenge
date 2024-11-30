package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type FileSpec struct {
	Name  string `yaml:"name"`
	Size  string `yaml:"size"`
	Count int    `yaml:"count,omitempty"`
}

type DirectorySpec struct {
	Name    string          `yaml:"name"`
	Files   []FileSpec      `yaml:"files"`
	Folders []DirectorySpec `yaml:"folders"`
}

func parseSize(sizeStr string) (int64, error) {
	units := map[string]int64{
		"B":  1,
		"KB": 1000,
		"MB": 1000 * 1000,
		"GB": 1000 * 1000 * 1000,
		"TB": 1000 * 1000 * 1000 * 1000,
	}

	sizeStr = strings.ToUpper(sizeStr)
	var unit string
	var value int64
	var err error

	for u := range units {
		if matched, _ := regexp.MatchString(`\d+\s*`+u+`$`, sizeStr); matched {
			// Extract the numeric part and the unit
			numericPart := strings.TrimSuffix(sizeStr, u)
			numericPart = strings.TrimSpace(numericPart)
			if value, err = strconv.ParseInt(numericPart, 10, 64); err != nil {
				return 0, err
			}
			unit = u
			break
		}
	}

	if unit == "" {
		// Handle case without unit (assuming bytes)
		if value, err = strconv.ParseInt(sizeStr, 10, 64); err != nil {
			return 0, fmt.Errorf("error parsing size string as integer: %w", err)
		}
	}

	return value * units[unit], nil
}

func createFile(path string, size int64) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return file.Truncate(size)
}

func createStructureFromSpec(spec DirectorySpec, parentPath string) error {
	currentPath := filepath.Join(parentPath, spec.Name)
	if err := os.MkdirAll(currentPath, 0700); err != nil {
		return err
	}

	for _, file := range spec.Files {
		size, err := parseSize(file.Size)
		if err != nil {
			return err
		}
		count := 1
		if file.Count > 1 {
			count = file.Count
		}
		for i := 0; i < count; i++ {
			fileName := file.Name
			if fileName == "" {
				fileName = fmt.Sprintf("%s file", file.Size)
			}
			if count > 1 {
				fileName = fmt.Sprintf("%s_%03d", fileName, i+1)
			}
			if err := createFile(filepath.Join(currentPath, fileName), size); err != nil {
				return err
			}
		}
	}

	for _, folder := range spec.Folders {
		if err := createStructureFromSpec(folder, currentPath); err != nil {
			return err
		}
	}

	return nil
}

func processYAMLFile(yamlFilePath string) error {
	data, err := os.ReadFile(yamlFilePath)
	if err != nil {
		return err
	}

	var root DirectorySpec
	if err := yaml.Unmarshal(data, &root); err != nil {
		return err
	}

	// Delete the root directory if it exists
	rootPath := filepath.Join("dist", root.Name)
	if err := os.RemoveAll(rootPath); err != nil {
		return err
	}

	return createStructureFromSpec(root, "dist")
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <yamlfile1> [<yamlfile2>...]", os.Args[0])
	}

	for _, yamlFile := range os.Args[1:] {
		if err := processYAMLFile(yamlFile); err != nil {
			log.Fatalf("Error processing '%s': %v", yamlFile, err)
		}
		log.Printf("Directory structure created successfully for '%s'.\n", yamlFile)
	}
}
