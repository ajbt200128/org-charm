package org

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	goorg "github.com/niklasfasching/go-org/org"
)

// FileEntry represents a file or directory in the file tree
type FileEntry struct {
	Name     string       // Display name
	Path     string       // Full path
	RelPath  string       // Relative path from root
	IsDir    bool         // Is this a directory?
	Parent   *FileEntry   // Parent directory (nil for root entries)
	Children []*FileEntry // Child entries (for directories)
	OrgFile  *OrgFile     // Parsed org file (for .org files)
	Expanded bool         // Is directory expanded in view?
}

// OrgFile represents a parsed org file
type OrgFile struct {
	Name       string
	Path       string
	Document   *goorg.Document
	RawContent string
}

// Title returns the document title from #+TITLE: or the filename
func (f *OrgFile) Title() string {
	if title := f.Document.Get("TITLE"); title != "" {
		return title
	}
	return strings.TrimSuffix(f.Name, ".org")
}

// Author returns the document author from #+AUTHOR:
func (f *OrgFile) Author() string {
	return f.Document.Get("AUTHOR")
}

// Date returns the document date from #+DATE:
func (f *OrgFile) Date() string {
	return f.Document.Get("DATE")
}

// ParseFile reads and parses an org file using go-org
func ParseFile(path string) (*OrgFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := goorg.New()
	doc := config.Parse(strings.NewReader(string(content)), path)

	return &OrgFile{
		Name:       filepath.Base(path),
		Path:       path,
		Document:   doc,
		RawContent: string(content),
	}, nil
}

// ListOrgFiles returns all .org files in a directory (non-recursive, for backwards compatibility)
func ListOrgFiles(dir string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".org") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}

// BuildFileTree recursively builds a file tree from the given directory
func BuildFileTree(rootDir string) ([]*FileEntry, error) {
	return buildFileTreeRecursive(rootDir, rootDir, nil)
}

func buildFileTreeRecursive(rootDir, currentDir string, parent *FileEntry) ([]*FileEntry, error) {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, err
	}

	var result []*FileEntry
	var dirs []*FileEntry
	var files []*FileEntry

	for _, entry := range entries {
		// Skip hidden files/directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(currentDir, entry.Name())
		relPath, _ := filepath.Rel(rootDir, fullPath)

		fe := &FileEntry{
			Name:    entry.Name(),
			Path:    fullPath,
			RelPath: relPath,
			IsDir:   entry.IsDir(),
			Parent:  parent,
		}

		if entry.IsDir() {
			// Recursively process directory
			children, err := buildFileTreeRecursive(rootDir, fullPath, fe)
			if err != nil {
				continue // Skip directories we can't read
			}
			// Only include directories that have org files (directly or in subdirs)
			if hasOrgFiles(children) {
				fe.Children = children
				fe.Expanded = false
				dirs = append(dirs, fe)
			}
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".org") {
			files = append(files, fe)
		}
	}

	// Sort directories and files alphabetically
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	// Directories first, then files
	result = append(result, dirs...)
	result = append(result, files...)

	return result, nil
}

// hasOrgFiles checks if a list of entries contains any org files (directly or in subdirs)
func hasOrgFiles(entries []*FileEntry) bool {
	for _, e := range entries {
		if !e.IsDir {
			return true
		}
		if e.IsDir && len(e.Children) > 0 {
			return true
		}
	}
	return false
}

// FlattenTree returns a flat list of visible entries based on expansion state
func FlattenTree(entries []*FileEntry) []*FileEntry {
	var result []*FileEntry
	flattenRecursive(entries, &result, 0)
	return result
}

func flattenRecursive(entries []*FileEntry, result *[]*FileEntry, depth int) {
	for _, e := range entries {
		*result = append(*result, e)
		if e.IsDir && e.Expanded {
			flattenRecursive(e.Children, result, depth+1)
		}
	}
}

// GetDepth returns the nesting depth of a file entry
func (fe *FileEntry) GetDepth() int {
	depth := 0
	p := fe.Parent
	for p != nil {
		depth++
		p = p.Parent
	}
	return depth
}

// GetOrgFile returns the parsed org file, parsing it on first access
func (fe *FileEntry) GetOrgFile() (*OrgFile, error) {
	if fe.IsDir {
		return nil, nil
	}
	if fe.OrgFile != nil {
		return fe.OrgFile, nil
	}
	orgFile, err := ParseFile(fe.Path)
	if err != nil {
		return nil, err
	}
	fe.OrgFile = orgFile
	return fe.OrgFile, nil
}
