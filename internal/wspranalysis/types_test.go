package wspranalysis

import (
	"testing"
	"time"
)

// TestReceptionReportTime tests the Time() method of ReceptionReport.
func TestReceptionReportTime(t *testing.T) {
	tests := []struct {
		name      string
		timeStr   string
		wantValid bool
	}{
		{
			name:      "valid datetime format",
			timeStr:   "2024-12-14 15:30:45",
			wantValid: true,
		},
		{
			name:      "invalid datetime format",
			timeStr:   "invalid",
			wantValid: false,
		},
		{
			name:      "empty string",
			timeStr:   "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReceptionReport{TimeStr: tt.timeStr}
			result := r.Time()

			if tt.wantValid && result.IsZero() {
				t.Errorf("Time() expected valid time, got zero value")
			}
			if !tt.wantValid && !result.IsZero() {
				t.Errorf("Time() expected zero value for invalid input, got %v", result)
			}
		})
	}
}

// TestReceptionReportTime_ValidParse tests that Time() correctly parses a valid datetime.
func TestReceptionReportTime_ValidParse(t *testing.T) {
	expectedStr := "2024-12-14 15:30:45"
	r := &ReceptionReport{TimeStr: expectedStr}

	result := r.Time()
	expectedTime, _ := time.Parse(time.DateTime, expectedStr)

	if result != expectedTime {
		t.Errorf("Time() = %v, want %v", result, expectedTime)
	}
}

// TestSnrNorm_dB tests the SnrNorm_dB method.
func TestSnrNorm_dB(t *testing.T) {
	tests := []struct {
		name            string
		snr_dB          int8
		power_dBm       int8
		txRefPower_dBm  int8
		expectedSnrNorm int8
	}{
		{
			name:            "basic calculation",
			snr_dB:          -10,
			power_dBm:       10,
			txRefPower_dBm:  43,
			expectedSnrNorm: 23, // -10 + 43 - 10 = 23
		},
		{
			name:            "negative values",
			snr_dB:          -20,
			power_dBm:       -5,
			txRefPower_dBm:  20,
			expectedSnrNorm: 5, // -20 + 20 - (-5) = 5
		},
		{
			name:            "equal power and ref",
			snr_dB:          -5,
			power_dBm:       30,
			txRefPower_dBm:  30,
			expectedSnrNorm: -5, // -5 + 30 - 30 = -5
		},
		{
			name:            "zero snr",
			snr_dB:          0,
			power_dBm:       20,
			txRefPower_dBm:  43,
			expectedSnrNorm: 23, // 0 + 43 - 20 = 23
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReceptionReport{
				Snr_dB:    tt.snr_dB,
				Power_dBm: tt.power_dBm,
			}

			result := r.SnrNorm_dB(tt.txRefPower_dBm)
			if result != tt.expectedSnrNorm {
				t.Errorf("SnrNorm_dB(%d) = %d, want %d", tt.txRefPower_dBm, result, tt.expectedSnrNorm)
			}
		})
	}
}

// TestBandNameToCode tests the BandNameToCode function.
func TestBandNameToCode(t *testing.T) {
	tests := []struct {
		name      string
		bandName  string
		wantCode  int
		wantError bool
	}{
		{
			name:      "160m band",
			bandName:  "160m",
			wantCode:  1,
			wantError: false,
		},
		{
			name:      "80m band",
			bandName:  "80m",
			wantCode:  3,
			wantError: false,
		},
		{
			name:      "20m band",
			bandName:  "20m",
			wantCode:  14,
			wantError: false,
		},
		{
			name:      "mf band",
			bandName:  "mf",
			wantCode:  0,
			wantError: false,
		},
		{
			name:      "lf band",
			bandName:  "lf",
			wantCode:  -1,
			wantError: false,
		},
		{
			name:      "invalid band",
			bandName:  "invalid",
			wantCode:  0,
			wantError: true,
		},
		{
			name:      "empty string",
			bandName:  "",
			wantCode:  0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := BandNameToCode(tt.bandName)

			if (err != nil) != tt.wantError {
				t.Errorf("BandNameToCode(%q) error = %v, wantError %v", tt.bandName, err, tt.wantError)
			}

			if !tt.wantError && code != tt.wantCode {
				t.Errorf("BandNameToCode(%q) = %d, want %d", tt.bandName, code, tt.wantCode)
			}
		})
	}
}

// TestBandNames tests the BandNames function.
func TestBandNames(t *testing.T) {
	names := BandNames()

	if len(names) == 0 {
		t.Error("BandNames() returned empty list")
	}

	// Check that all expected bands are present
	expectedBands := map[string]bool{
		"lf": false, "mf": false, "160m": false, "80m": false,
		"60m": false, "40m": false, "30m": false, "20m": false,
		"17m": false, "15m": false, "12m": false, "10m": false,
		"6m": false, "4m": false, "2m": false, "70cm": false, "23cm": false,
	}

	for _, name := range names {
		if _, ok := expectedBands[name]; !ok {
			t.Errorf("BandNames() returned unexpected band: %s", name)
		}
		expectedBands[name] = true
	}

	for band, found := range expectedBands {
		if !found {
			t.Errorf("BandNames() missing expected band: %s", band)
		}
	}
}

// TestBandNamesCaseInsensitivity tests BandNameToCode with different cases.
func TestBandNamesCaseInsensitivity(t *testing.T) {
	// BandNameToCode expects lowercase band names
	tests := []struct {
		bandName string
		wantCode int
	}{
		{"20m", 14},
		{"2M", 144},
	}

	for _, tt := range tests {
		code, err := BandNameToCode(tt.bandName)
		if err != nil {
			t.Errorf("BandNameToCode(%s) unexpected error: %v", tt.bandName, err)
		}
		if code != tt.wantCode {
			t.Errorf("BandNameToCode(%s) = %d, want %d", tt.bandName, code, tt.wantCode)
		}
	}
}
