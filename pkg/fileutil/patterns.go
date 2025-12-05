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
func FindAIFiles(repoPath string, patterns []string) ([]string, error) {
	var results []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// Clean the pattern
		cleanPattern := strings.TrimSuffix(pattern, "/")
		cleanPattern = strings.TrimSuffix(cleanPattern, "/**")

		// Check if the path exists directly
		fullPath := filepath.Join(repoPath, cleanPattern)
		if _, err := os.Stat(fullPath); err == nil {
			if !seen[cleanPattern] {
				results = append(results, cleanPattern)
				seen[cleanPattern] = true
			}
		}
	}

	return results, nil
}

// ExpandPatterns expands patterns to actual paths in the repository
// Returns a map of relative path -> full path
func ExpandPatterns(repoPath string, patterns []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, pattern := range patterns {
		// Clean the pattern (remove trailing / and /**)
		cleanPattern := strings.TrimSuffix(pattern, "/")
		cleanPattern = strings.TrimSuffix(cleanPattern, "/**")

		// Skip patterns that are just wildcards
		if cleanPattern == "" || cleanPattern == "*" || cleanPattern == "**" {
			continue
		}

		fullPath := filepath.Join(repoPath, cleanPattern)

		// Check if path exists
		if _, err := os.Stat(fullPath); err == nil {
			result[cleanPattern] = fullPath
		}
	}

	return result, nil
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
