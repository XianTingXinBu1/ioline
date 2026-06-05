package search

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	maxFileSearchResults = 100
	maxTextSearchResults = 200
	maxTextFileSize      = 1 << 20 // 1 MiB
)

var ErrQueryRequired = errors.New("query is required")

var ignoredDirectoryNames = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	".runtime":     {},
	".tmp":         {},
}

type FileResult struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

type FileSearchResult struct {
	Query string       `json:"query"`
	Items []FileResult `json:"items"`
	Total int          `json:"total"`
	Limit int          `json:"limit"`
}

type TextSearchRequest struct {
	Query string `json:"query"`
}

type TextMatch struct {
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	LineText  string `json:"lineText"`
	MatchText string `json:"matchText"`
}

type TextSearchResult struct {
	Query string      `json:"query"`
	Items []TextMatch `json:"items"`
	Total int         `json:"total"`
	Limit int         `json:"limit"`
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) FindFiles(rootPath, query string) (FileSearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return FileSearchResult{}, ErrQueryRequired
	}

	needle := strings.ToLower(query)
	items := make([]FileResult, 0, minInt(16, maxFileSearchResults))
	total := 0

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == rootPath {
			return nil
		}
		if d.IsDir() && shouldIgnoreDir(d.Name()) {
			return filepath.SkipDir
		}

		name := d.Name()
		if !strings.Contains(strings.ToLower(name), needle) {
			return nil
		}

		total++
		if len(items) >= maxFileSearchResults {
			return nil
		}

		kind := "file"
		if d.IsDir() {
			kind = "directory"
		}

		relPath, relErr := filepath.Rel(rootPath, path)
		if relErr != nil {
			return relErr
		}
		items = append(items, FileResult{
			Name: name,
			Path: filepath.ToSlash(relPath),
			Type: kind,
		})
		return nil
	})
	if err != nil {
		return FileSearchResult{}, err
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == "directory"
		}
		return strings.ToLower(items[i].Path) < strings.ToLower(items[j].Path)
	})

	return FileSearchResult{Query: query, Items: items, Total: total, Limit: maxFileSearchResults}, nil
}

func (s *Service) FindText(rootPath string, req TextSearchRequest) (TextSearchResult, error) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return TextSearchResult{}, ErrQueryRequired
	}

	needle := strings.ToLower(query)
	items := make([]TextMatch, 0, minInt(16, maxTextSearchResults))
	total := 0

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == rootPath {
			return nil
		}
		if d.IsDir() {
			if shouldIgnoreDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Size() > maxTextFileSize || info.Mode()&0o111 != 0 {
			return nil
		}

		matches, err := findTextMatchesInFile(rootPath, path, needle, query, maxTextSearchResults-len(items), &total)
		if err != nil {
			return nil
		}
		items = append(items, matches...)
		return nil
	})
	if err != nil {
		return TextSearchResult{}, err
	}

	return TextSearchResult{Query: query, Items: items, Total: total, Limit: maxTextSearchResults}, nil
}

func findTextMatchesInFile(rootPath, absPath, needle, rawQuery string, remaining int, total *int) ([]TextMatch, error) {
	if remaining <= 0 {
		return nil, nil
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, maxTextFileSize)

	matches := make([]TextMatch, 0, minInt(4, remaining))
	lineNumber := 0
	binaryChecked := false

	relPath, err := filepath.Rel(rootPath, absPath)
	if err != nil {
		return nil, err
	}
	relPath = filepath.ToSlash(relPath)

	for scanner.Scan() {
		lineNumber++
		lineText := scanner.Text()
		if !binaryChecked {
			binaryChecked = true
			if isBinary([]byte(lineText)) {
				return nil, nil
			}
		}

		lowerLine := strings.ToLower(lineText)
		searchFrom := 0
		for {
			idx := strings.Index(lowerLine[searchFrom:], needle)
			if idx < 0 {
				break
			}
			actualIndex := searchFrom + idx
			*total++
			if len(matches) < remaining {
				matches = append(matches, TextMatch{
					Path:      relPath,
					Line:      lineNumber,
					Column:    actualIndex + 1,
					LineText:  lineText,
					MatchText: lineText[actualIndex : actualIndex+len(rawQuery)],
				})
			}
			searchFrom = actualIndex + len(needle)
			if searchFrom >= len(lowerLine) {
				break
			}
		}
		if len(matches) >= remaining {
			break
		}
	}

	return matches, scanner.Err()
}

func shouldIgnoreDir(name string) bool {
	_, ok := ignoredDirectoryNames[name]
	return ok
}

func isBinary(content []byte) bool {
	for _, b := range content {
		if b == 0 {
			return true
		}
	}
	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
