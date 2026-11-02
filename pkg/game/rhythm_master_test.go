package game

import "testing"

func newRhythmMasterGame(t *testing.T) *Game {
	t.Helper()
	g := New()
	if err := g.LoadWordDictionaries(
		"../../data/google-10000-short.txt",
		"../../data/google-10000-medium.txt",
		"../../data/google-10000-long.txt",
		0.3, 0.5, 0.2,
	); err != nil {
		t.Fatalf("LoadWordDictionaries: %v", err)
	}
	return g
}

// eliminateFirstWord types and submits the current word
func eliminateFirstWord(g *Game) {
	for _, w := range g.Words {
		if !w.Completed {
			for _, r := range w.Text {
				g.AddChar(r)
			}
			g.TryEliminate()
			return
		}
	}
}

// TestRhythmMasterZeroWordsPerLevel used to panic with integer divide by zero
func TestRhythmMasterZeroWordsPerLevel(t *testing.T) {
	g := newRhythmMasterGame(t)
	g.RhythmInitialTimeLimit = 60
	g.RhythmWordsPerLevel = 0
	if err := g.StartRhythmMasterMode(); err != nil {
		t.Fatal(err)
	}

	eliminateFirstWord(g)
	if g.ConsecutiveSuccesses != 1 || g.DifficultyLevel != 0 {
		t.Errorf("successes=%d level=%d", g.ConsecutiveSuccesses, g.DifficultyLevel)
	}
}

func TestRhythmMasterLevelsUp(t *testing.T) {
	g := newRhythmMasterGame(t)
	g.RhythmInitialTimeLimit = 60
	g.RhythmMinTimeLimit = 59.85
	g.RhythmDifficultyStep = 0.1
	g.RhythmWordsPerLevel = 2
	if err := g.StartRhythmMasterMode(); err != nil {
		t.Fatal(err)
	}

	eliminateFirstWord(g)
	eliminateFirstWord(g)
	if g.DifficultyLevel != 1 || g.WordTimeLimit.Seconds() != 59.9 {
		t.Errorf("level=%d limit=%v", g.DifficultyLevel, g.WordTimeLimit)
	}

	eliminateFirstWord(g)
	eliminateFirstWord(g)
	if g.DifficultyLevel != 2 || g.WordTimeLimit.Seconds() != 59.85 {
		t.Errorf("limit should clamp to the minimum: level=%d limit=%v", g.DifficultyLevel, g.WordTimeLimit)
	}
}
