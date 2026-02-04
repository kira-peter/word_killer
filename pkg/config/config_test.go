package config

import (
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
