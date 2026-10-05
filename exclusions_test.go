package main

import "testing"

func TestMergeExclusionsAlwaysKeepsDefaults(t *testing.T) {
	merged := mergeExclusions([]string{"/reports", "/_health"})
	found := map[string]bool{}
	for _, entry := range merged {
		found[entry] = true
	}
	for _, want := range []string{"/_health", "/_ready", "/docs", "/reports"} {
		if !found[want] {
			t.Fatalf("merged exclusions must contain %q, got %v", want, merged)
		}
	}
	if len(merged) != len(found) {
		t.Fatalf("duplicates must be dropped, got %v", merged)
	}
}

func TestPathExclusionsSubtreeSemantics(t *testing.T) {
	p := newPathExclusions([]string{"/health", "", "/reports/"})
	if !p.matches("/health") {
		t.Fatal("equal path must match")
	}
	if !p.matches("/health/ready") {
		t.Fatal("subtree must match")
	}
	if p.matches("/healthcheck") {
		t.Fatal("a sibling prefix must not match")
	}
	if !p.matches("/reports/weekly/x") {
		t.Fatal("trailing-slash entry must behave like a subtree")
	}
	if p.matches("/other") {
		t.Fatal("unrelated path must not match")
	}
}
