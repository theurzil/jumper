package main

import (
	"testing"
	"time"
)

func TestScore(t *testing.T) {
	now := time.Now().Unix()

	recent := Entity{Path: "/recent", Frequency: 1, LastVisit: now}
	old := Entity{Path: "/old", Frequency: 1, LastVisit: now - 8*86400}

	if recent.score() <= old.score() {
		t.Errorf("expected recent entry to score higher than old entry: recent=%v old=%v", recent.score(), old.score())
	}
}

func TestScoreHigherFrequencyWins(t *testing.T) {
	now := time.Now().Unix()

	frequent := Entity{Path: "/frequent", Frequency: 10, LastVisit: now}
	rare := Entity{Path: "/rare", Frequency: 1, LastVisit: now}

	if frequent.score() <= rare.score() {
		t.Errorf("expected higher frequency to score higher: frequent=%v rare=%v", frequent.score(), rare.score())
	}
}

func TestQueryPrefersExactBasenameOverFrequentSubstring(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	now := time.Now().Unix()
	entities := Entities{
		{Path: "/home/u/dev/icc", Frequency: 1, LastVisit: now},
		{Path: "/home/u/dev/icc-back", Frequency: 50, LastVisit: now},
	}
	if err := saveEntities(entities); err != nil {
		t.Fatalf("saveEntities: %v", err)
	}

	got, err := query("icc")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got != "/home/u/dev/icc" {
		t.Errorf("expected exact basename match /home/u/dev/icc, got %s", got)
	}
}

func TestRankedMatchesOrdersBestFirst(t *testing.T) {
	now := time.Now().Unix()
	entities := Entities{
		{Path: "/home/u/dev/foo-bar", Frequency: 50, LastVisit: now},
		{Path: "/home/u/dev/foo", Frequency: 1, LastVisit: now},
		{Path: "/home/u/other", Frequency: 1, LastVisit: now},
	}

	matches := rankedMatches(entities, "foo")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d: %v", len(matches), matches)
	}
	if matches[0].Path != "/home/u/dev/foo" {
		t.Errorf("expected /home/u/dev/foo first, got %s", matches[0].Path)
	}
	if matches[1].Path != "/home/u/dev/foo-bar" {
		t.Errorf("expected /home/u/dev/foo-bar second, got %s", matches[1].Path)
	}
}

func TestQueryPrefersBaseDirOverVisitedSubdir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	now := time.Now().Unix()
	entities := Entities{
		{Path: "/home/u/github", Frequency: 1, LastVisit: now - 1000},
		{Path: "/home/u/github/some-project", Frequency: 5, LastVisit: now},
	}
	if err := saveEntities(entities); err != nil {
		t.Fatalf("saveEntities: %v", err)
	}

	got, err := query("github")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if got != "/home/u/github" {
		t.Errorf("expected base dir /home/u/github, got %s", got)
	}
}
