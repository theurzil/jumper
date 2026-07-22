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
