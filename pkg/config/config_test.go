package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeRatios(t *testing.T) {
	tests := []struct {
		name        string
		short       float64
		medium      float64
		long        float64
		wantShort   float64
		wantMedium  float64
		wantLong    float64
		expectError bool
	}{
		{
			name:        "Standard percentages",
			short:       30,
			medium:      50,
			long:        20,
			wantShort:   0.3,
			wantMedium:  0.5,
			wantLong:    0.2,
			expectError: false,
		},
		{
			name:        "Simple ratio 1:2:1",
			short:       1,
			medium:      2,
			long:        1,
			wantShort:   0.25,
			wantMedium:  0.5,
			wantLong:    0.25,
			expectError: false,
		},
		{
			name:        "One ratio is zero",
			short:       0,
			medium:      70,
			long:        30,
			wantShort:   0,
			wantMedium:  0.7,
			wantLong:    0.3,
			expectError: false,
		},
		{
			name:        "All ratios zero",
			short:       0,
			medium:      0,
			long:        0,
			expectError: true,
		},
		{
			name:        "Negative ratio",
			short:       -1,
			medium:      50,
			long:        50,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ShortRatio:  tt.short,
				MediumRatio: tt.medium,
				LongRatio:   tt.long,
			}

			gotShort, gotMedium, gotLong, err := cfg.NormalizeRatios()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Use approximate comparison for floats
			epsilon := 0.0001
			if abs(gotShort-tt.wantShort) > epsilon {
				t.Errorf("Short ratio = %v, want %v", gotShort, tt.wantShort)
			}
			if abs(gotMedium-tt.wantMedium) > epsilon {
				t.Errorf("Medium ratio = %v, want %v", gotMedium, tt.wantMedium)
			}
			if abs(gotLong-tt.wantLong) > epsilon {
				t.Errorf("Long ratio = %v, want %v", gotLong, tt.wantLong)
			}

			// Verify total is 1.0
			total := gotShort + gotMedium + gotLong
			if abs(total-1.0) > epsilon {
				t.Errorf("Total ratio = %v, want 1.0", total)
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestApplyOnlineDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.applyOnlineDefaults()

	if cfg.OnlinePort != DefaultOnlinePort {
		t.Errorf("OnlinePort = %d, want %d", cfg.OnlinePort, DefaultOnlinePort)
	}
	if cfg.OnlineDropDuration != DefaultOnlineDropDuration {
		t.Errorf("OnlineDropDuration = %d, want %d", cfg.OnlineDropDuration, DefaultOnlineDropDuration)
	}
	if cfg.OnlinePlayerName == "" {
		t.Error("OnlinePlayerName should not be empty")
	}
}

func TestApplyOnlineDefaultsKeepsCustomValues(t *testing.T) {
	cfg := &Config{OnlinePort: 9000, OnlineDropDuration: 5, OnlinePlayerName: "Alice"}
	cfg.applyOnlineDefaults()

	if cfg.OnlinePort != 9000 || cfg.OnlineDropDuration != 5 || cfg.OnlinePlayerName != "Alice" {
		t.Errorf("custom values changed: %+v", cfg)
	}
}

func TestPlayerNameTruncation(t *testing.T) {
	cfg := &Config{OnlinePlayerName: "AVeryLongPlayerName"}
	cfg.applyOnlineDefaults()

	if got := []rune(cfg.OnlinePlayerName); len(got) != MaxPlayerNameLength {
		t.Errorf("name length = %d, want %d", len(got), MaxPlayerNameLength)
	}
	if cfg.OnlinePlayerName != "AVeryLongPla" {
		t.Errorf("name = %q, want %q", cfg.OnlinePlayerName, "AVeryLongPla")
	}
}

func TestApplyRhythmMasterDefaults(t *testing.T) {
	defaults := DefaultConfig()

	cfg := &Config{}
	cfg.applyRhythmMasterDefaults()
	if cfg.RhythmInitialTimeLimit != defaults.RhythmInitialTimeLimit ||
		cfg.RhythmMinTimeLimit != defaults.RhythmMinTimeLimit ||
		cfg.RhythmWordsPerLevel != defaults.RhythmWordsPerLevel {
		t.Errorf("zero values not replaced: %+v", cfg)
	}

	// A zero step is valid: the time limit simply never shrinks
	if cfg.RhythmDifficultyStep != 0 {
		t.Errorf("zero step should be kept, got %v", cfg.RhythmDifficultyStep)
	}

	cfg = &Config{RhythmInitialTimeLimit: 1, RhythmMinTimeLimit: 3, RhythmDifficultyStep: -1, RhythmWordsPerLevel: -5}
	cfg.applyRhythmMasterDefaults()
	if cfg.RhythmMinTimeLimit != 1 || cfg.RhythmDifficultyStep != defaults.RhythmDifficultyStep || cfg.RhythmWordsPerLevel != defaults.RhythmWordsPerLevel {
		t.Errorf("invalid values not fixed: %+v", cfg)
	}

	cfg = &Config{RhythmInitialTimeLimit: 4, RhythmMinTimeLimit: 0.5, RhythmDifficultyStep: 0.2, RhythmWordsPerLevel: 5}
	cfg.applyRhythmMasterDefaults()
	if cfg.RhythmInitialTimeLimit != 4 || cfg.RhythmMinTimeLimit != 0.5 || cfg.RhythmDifficultyStep != 0.2 || cfg.RhythmWordsPerLevel != 5 {
		t.Errorf("valid values changed: %+v", cfg)
	}
}

func TestLoadAppliesRhythmMasterDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"rhythm_initial_time_limit": 0}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RhythmInitialTimeLimit <= 0 || cfg.RhythmWordsPerLevel <= 0 {
		t.Errorf("Load kept invalid rhythm settings: %+v", cfg)
	}
}
