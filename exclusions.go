package main

import "strings"

// pathExclusions holds normalized exclude_paths entries.
type pathExclusions struct {
	entries []string
}

// newPathExclusions drops empty entries, trims whitespace and trailing
// slashes, and keeps the rest as-is.
func newPathExclusions(entries []string) *pathExclusions {
	kept := make([]string, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimRight(strings.TrimSpace(entry), "/")
		if entry == "" {
			continue
		}
		kept = append(kept, entry)
	}
	return &pathExclusions{entries: kept}
}

// matches reports whether path falls inside one of the exclusion subtrees,
// using the same subtree-or-equal semantics as the engine's exclude path
// matcher: an entry "/health" covers "/health" and anything under "/health/".
func (p *pathExclusions) matches(path string) bool {
	for _, entry := range p.entries {
		if entry == "/" || path == entry || strings.HasPrefix(path, entry+"/") {
			return true
		}
	}
	return false
}

// mergeExclusions combines the built-in defaults (health probes and the
// engine's documentation/static paths) with operator entries, dropping
// duplicates. Defaults are always on: opting a health probe out of guard
// checks is how a shared-identity ban takes a pod out of rotation.
func mergeExclusions(operator []string) []string {
	seen := make(map[string]bool, len(defaultExcludedPaths)+len(operator))
	merged := make([]string, 0, len(defaultExcludedPaths)+len(operator))
	for _, entry := range append(append([]string(nil), defaultExcludedPaths...), operator...) {
		entry = strings.TrimSpace(entry)
		if entry == "" || seen[entry] {
			continue
		}
		seen[entry] = true
		merged = append(merged, entry)
	}
	return merged
}
