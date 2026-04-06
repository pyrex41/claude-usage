package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pyrex41/claude-usage/internal/types"
)

type Parser struct {
	Filters types.Filters
}

func (p *Parser) ParseFiles(paths []string, ch chan<- types.Event, wg *sync.WaitGroup) {
	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			p.parseFile(path, ch)
		}(path)
	}
}

func parseProjectName(name string) string {
	if name == "unknown" || name == "" {
		return "Unknown Project"
	}

	cleaned := name

	// Handle Unix-style paths: /Users/... or -Users-...
	if strings.HasPrefix(cleaned, "-Users-") || strings.HasPrefix(cleaned, "/Users/") {
		sep := "-"
		if strings.HasPrefix(cleaned, "/Users/") {
			sep = "/"
		}
		segments := strings.FieldsFunc(cleaned, func(r rune) bool {
			return r == rune(sep[0]) || r == '/'
		})
		for i, seg := range segments {
			if seg == "Users" && i+2 < len(segments) {
				cleaned = strings.Join(segments[i+2:], "-")
				break
			}
		}
	}

	// If no path cleanup, basic cleanup
	if cleaned == name {
		cleaned = strings.Trim(cleaned, "/\\-")
	}

	// Handle UUID-like patterns - use last 2 segments
	if isUUID(cleaned) {
		parts := strings.Split(cleaned, "-")
		if len(parts) >= 5 {
			cleaned = parts[len(parts)-2] + "-" + parts[len(parts)-1]
		}
	}

	// Handle project--branch patterns
	if idx := strings.Index(cleaned, "--"); idx > 0 {
		cleaned = cleaned[:idx]
	}

	// For compound names > 20 chars, try to extract meaningful part
	if len(cleaned) > 20 && strings.Contains(cleaned, "-") {
		segments := strings.Split(cleaned, "-")
		if len(segments) >= 2 {
			cleaned = strings.Join(segments[len(segments)-2:], "-")
		}
	}

	cleaned = strings.Trim(cleaned, "/\\-")
	if cleaned == "" {
		if name != "" {
			return name
		}
		return "Unknown Project"
	}
	return cleaned
}

func isUUID(s string) bool {
	s = strings.TrimSuffix(s, ".jsonl")
	parts := strings.Split(s, "-")
	if len(parts) != 5 {
		return false
	}
	for _, p := range parts {
		if len(p) != 4 && len(p) != 12 {
			return false
		}
		for _, c := range p {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

func (p *Parser) parseFile(path string, ch chan<- types.Event) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	dir := filepath.Dir(path)
	parts := strings.Split(dir, string(os.PathSeparator))

	// Extract project from path
	proj := ""
	sessionDir := ""
	for i, part := range parts {
		if part == "projects" && i+1 < len(parts) {
			proj = parseProjectName(parts[i+1])
			// Check if there's a session subdirectory
			if i+2 < len(parts) {
				sessionDir = parts[i+2]
			}
			break
		}
	}
	if proj == "" {
		proj = "Unknown Project"
	}

	// For files directly in project folder (no session subdir), use project name as session
	// For files in session subdirs, use session name (or project if UUID)
	inst := filepath.Base(path)
	inst = strings.TrimSuffix(inst, ".jsonl")
	if isUUID(inst) || inst == proj || sessionDir == "" {
		inst = proj
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		dec := json.NewDecoder(bytes.NewReader(line))
		var event types.Event
		if err := dec.Decode(&event); err != nil {
			continue
		}
		t := event.Time()
		if t.IsZero() || t.Year() < 2020 {
			continue
		}
		if !p.filter(t, proj) {
			continue
		}
		event.Project = proj
		event.Instance = inst
		select {
		case ch <- event:
		default:
		}
	}
}

func (p *Parser) filter(t time.Time, proj string) bool {
	if p.Filters.Since != "" {
		since, _ := time.Parse("20060102", p.Filters.Since)
		if t.Before(since) {
			return false
		}
	}
	if p.Filters.Until != "" {
		until, _ := time.Parse("20060102", p.Filters.Until)
		if t.After(until.Add(24*time.Hour - time.Second)) {
			return false
		}
	}
	if p.Filters.Project != "" && proj != p.Filters.Project {
		return false
	}
	return true
}

func FindFiles(basePath string) ([]string, error) {
	var paths []string
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".jsonl") {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}
