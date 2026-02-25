package org

import (
	"os"
	"path/filepath"
	"strings"

	goorg "github.com/niklasfasching/go-org/org"
)

// OrgFile represents a parsed org file
type OrgFile struct {
	Name     string
	Path     string
	Document *goorg.Document
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
		Name:     filepath.Base(path),
		Path:     path,
		Document: doc,
	}, nil
}

// ListOrgFiles returns all .org files in a directory
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
