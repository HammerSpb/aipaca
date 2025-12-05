package fileutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// MatchPatterns finds all files/directories in baseDir that match any of the patterns
func MatchPatterns(baseDir string, patterns []string) ([]string, error) {
	var matches []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Handle patterns that match directories (e.g., ".claude", "ai/")
		fullPattern := filepath.Join(baseDir, pattern)

		// Check if the pattern ends with / (directory pattern)
		isDirectoryPattern := strings.HasSuffix(pattern, "/")
		cleanPattern := strings.TrimSuffix(pattern, "/")

		// First, check if the pattern matches a directory directly
		dirPath := filepath.Join(baseDir, cleanPattern)
		if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
			if !seen[cleanPattern] {
				matches = append(matches, cleanPattern)
				seen[cleanPattern] = true
			}
			continue
		}

		// If it's a directory pattern but not a wildcard, skip file matching
		if isDirectoryPattern && !strings.Contains(pattern, "*") {
			continue
		}

		// Use doublestar for glob matching
		matched, err := doublestar.FilepathGlob(fullPattern)
		if err != nil {
			return nil, err
		}

		for _, m := range matched {
			relPath, err := filepath.Rel(baseDir, m)
			if err != nil {
				continue
			}
			// Skip if already seen
			if seen[relPath] {
				continue
			}
			// For ** patterns, we want the top-level item
			topLevel := getTopLevelPath(relPath)
			if !seen[topLevel] {
				matches = append(matches, topLevel)
				seen[topLevel] = true
			}
		}
	}

	return matches, nil
}

// getTopLevelPath returns the first component of a path
func getTopLevelPath(path string) string {
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	if len(parts) > 0 {
		return parts[0]
	}
	return path
}

// FindAIFiles finds all AI-related files/directories in a repo based on patterns
// Returns top-level items: directories for "dir/**" patterns, individual files for "**/file" patterns
func FindAIFiles(repoPath string, patterns []string) ([]string, error) {
	var results []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Skip empty patterns
		if pattern == "" {
			continue
		}

		// Check if this is a "directory contents" pattern (ends with /**)
		if strings.HasSuffix(pattern, "/**") {
			dirPattern := strings.TrimSuffix(pattern, "/**")
			fullPath := filepath.Join(repoPath, dirPattern)
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				if !seen[dirPattern] {
					results = append(results, dirPattern)
					seen[dirPattern] = true
				}
			}
			continue
		}

		// Check if pattern contains wildcards (like **/CLAUDE.md)
		if strings.Contains(pattern, "*") {
			fullPattern := filepath.Join(repoPath, pattern)
			matched, err := doublestar.FilepathGlob(fullPattern)
			if err != nil {
				continue
			}
			for _, m := range matched {
				relPath, err := filepath.Rel(repoPath, m)
				if err != nil {
					continue
				}
				// Skip if under an already-seen directory
				skip := false
				for seenPath := range seen {
					if strings.HasPrefix(relPath, seenPath+string(filepath.Separator)) {
						skip = true
						break
					}
				}
				if !skip && !seen[relPath] {
					results = append(results, relPath)
					seen[relPath] = true
				}
			}
		} else {
			// Clean the pattern (remove trailing /)
			cleanPattern := strings.TrimSuffix(pattern, "/")

			// Check if the path exists directly
			fullPath := filepath.Join(repoPath, cleanPattern)
			if _, err := os.Stat(fullPath); err == nil {
				if !seen[cleanPattern] {
					results = append(results, cleanPattern)
					seen[cleanPattern] = true
				}
			}
		}
	}

	return results, nil
}

// ExpandPatterns expands patterns to actual paths in the repository
// Returns a map of relative path -> full path
// For directory patterns like ".claude/**", returns the directory itself
// For file patterns like "**/CLAUDE.md", returns matched files
func ExpandPatterns(repoPath string, patterns []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, pattern := range patterns {
		// Skip patterns that are just wildcards
		if pattern == "" || pattern == "*" || pattern == "**" {
			continue
		}

		// Check if this is a "directory contents" pattern (ends with /**)
		if strings.HasSuffix(pattern, "/**") {
			// For patterns like ".claude/**", just check if the directory exists
			dirPattern := strings.TrimSuffix(pattern, "/**")
			fullPath := filepath.Join(repoPath, dirPattern)
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				result[dirPattern] = fullPath
			}
			continue
		}

		// Check if pattern contains wildcards (like **/CLAUDE.md)
		if strings.Contains(pattern, "*") {
			// Use doublestar for glob matching
			fullPattern := filepath.Join(repoPath, pattern)
			matched, err := doublestar.FilepathGlob(fullPattern)
			if err != nil {
				continue
			}
			for _, m := range matched {
				relPath, err := filepath.Rel(repoPath, m)
				if err != nil {
					continue
				}
				// Skip if parent directory is already in result
				if !isUnderExistingPath(relPath, result) {
					result[relPath] = m
				}
			}
		} else {
			// Clean the pattern (remove trailing /)
			cleanPattern := strings.TrimSuffix(pattern, "/")
			fullPath := filepath.Join(repoPath, cleanPattern)

			// Check if path exists
			if _, err := os.Stat(fullPath); err == nil {
				result[cleanPattern] = fullPath
			}
		}
	}

	return result, nil
}

// isUnderExistingPath checks if path is under any existing path in the result map
func isUnderExistingPath(path string, existing map[string]string) bool {
	for existingPath := range existing {
		if strings.HasPrefix(path, existingPath+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// IsAIFile checks if a path matches any AI file pattern
func IsAIFile(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple check for direct matches or prefix matches
		cleanPattern := strings.TrimSuffix(pattern, "/")
		cleanPattern = strings.TrimSuffix(cleanPattern, "/**")

		if strings.HasPrefix(path, cleanPattern) {
			return true
		}

		// Check glob pattern
		matched, _ := doublestar.Match(pattern, path)
		if matched {
			return true
		}
	}
	return false
}
