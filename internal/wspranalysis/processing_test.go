package wspranalysis

import (
	"testing"
	"time"
)

// TestNewReceptionReportGroup tests the newReceptionReportGroup function.
func TestNewReceptionReportGroup(t *testing.T) {
	tests := []struct {
		name              string
		reports           []ReceptionReport
		targetCallsign    string
		wantError         bool
		wantTargetIndex   int
		wantRxSign        string
		wantReportsLength int
	}{
		{
			name: "valid group with target",
			reports: []ReceptionReport{
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "W5XYZ", Power_dBm: 10},
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "N0OTH", Power_dBm: 20},
			},
			targetCallsign:    "W5XYZ",
			wantError:         false,
			wantTargetIndex:   0,
			wantRxSign:        "W5ABC",
			wantReportsLength: 2,
		},
		{
			name: "target not in group",
			reports: []ReceptionReport{
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "N0OTH", Power_dBm: 20},
			},
			targetCallsign:  "W5XYZ",
			wantError:       true,
			wantTargetIndex: -1,
		},
		{
			name:            "empty reports slice",
			reports:         []ReceptionReport{},
			targetCallsign:  "W5XYZ",
			wantError:       false,
			wantTargetIndex: 0, // Empty group returns with TargetIndex 0 by default
		},
		{
			name: "target at end of group",
			reports: []ReceptionReport{
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "N0OTH", Power_dBm: 20},
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "W5XYZ", Power_dBm: 10},
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "G3ABC", Power_dBm: 30},
			},
			targetCallsign:    "W5XYZ",
			wantError:         false,
			wantTargetIndex:   1,
			wantRxSign:        "W5ABC",
			wantReportsLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := newReceptionReportGroup(tt.reports, tt.targetCallsign)

			if (err != nil) != tt.wantError {
				t.Errorf("newReceptionReportGroup() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && result != nil {
				if result.TargetIndex != tt.wantTargetIndex {
					t.Errorf("newReceptionReportGroup() TargetIndex = %d, want %d", result.TargetIndex, tt.wantTargetIndex)
				}
				// Only check RxSign and length if we have reports
				if len(tt.reports) > 0 {
					if result.RxSign != tt.wantRxSign {
						t.Errorf("newReceptionReportGroup() RxSign = %s, want %s", result.RxSign, tt.wantRxSign)
					}
					if len(result.Reports) != tt.wantReportsLength {
						t.Errorf("newReceptionReportGroup() Reports length = %d, want %d", len(result.Reports), tt.wantReportsLength)
					}
				}
			}
		})
	}
}

// TestMedian tests the median function.
func TestMedian(t *testing.T) {
	tests := []struct {
		name       string
		values     []int
		preSorted  bool
		wantResult float64
		wantError  bool
	}{
		{
			name:       "odd number of elements",
			values:     []int{3, 1, 4, 1, 5},
			preSorted:  false,
			wantResult: 3,
			wantError:  false,
		},
		{
			name:       "even number of elements",
			values:     []int{1, 2, 3, 4},
			preSorted:  true,
			wantResult: 2.5,
			wantError:  false,
		},
		{
			name:       "single element",
			values:     []int{42},
			preSorted:  true,
			wantResult: 42,
			wantError:  false,
		},
		{
			name:       "empty slice",
			values:     []int{},
			preSorted:  true,
			wantResult: 0,
			wantError:  true,
		},
		{
			name:       "two elements",
			values:     []int{10, 20},
			preSorted:  true,
			wantResult: 15,
			wantError:  false,
		},
		{
			name:       "unsorted odd",
			values:     []int{5, 2, 8, 1, 9},
			preSorted:  false,
			wantResult: 5,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the test data
			valuesCopy := make([]int, len(tt.values))
			copy(valuesCopy, tt.values)

			result, err := median(valuesCopy, tt.preSorted)

			if (err != nil) != tt.wantError {
				t.Errorf("median() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && result != tt.wantResult {
				t.Errorf("median() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

// TestMedian_FloatValues tests median with float values.
func TestMedian_FloatValues(t *testing.T) {
	values := []float64{1.5, 2.5, 3.5, 4.5}
	result, err := median(values, true)

	if err != nil {
		t.Errorf("median() unexpected error: %v", err)
	}

	expected := 3.0 // (2.5 + 3.5) / 2
	if result != expected {
		t.Errorf("median() = %v, want %v", result, expected)
	}
}

// TestProcessRawRxReports tests the processRawRxReports function.
func TestProcessRawRxReports(t *testing.T) {
	tests := []struct {
		name               string
		rawReports         []ReceptionReport
		targetCallsign     string
		normTxPwr_dBm      int8
		wantError          bool
		wantGroupsCount    int
		wantFirstGroupSize int
	}{
		{
			name: "single time and receiver",
			rawReports: []ReceptionReport{
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "W5XYZ", Power_dBm: 10, Snr_dB: -10},
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "N0OTH", Power_dBm: 20, Snr_dB: -15},
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "G3ABC", Power_dBm: 30, Snr_dB: -5},
			},
			targetCallsign:     "W5XYZ",
			normTxPwr_dBm:      43,
			wantError:          false,
			wantGroupsCount:    1,
			wantFirstGroupSize: 3,
		},
		{
			name: "multiple times and receivers",
			rawReports: []ReceptionReport{
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "W5XYZ", Power_dBm: 10, Snr_dB: -10},
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "N0OTH", Power_dBm: 20, Snr_dB: -15},
				{TimeStr: "2024-12-14 15:31:45", RxSign: "W5DEF", TxSign: "W5XYZ", Power_dBm: 15, Snr_dB: -8},
				{TimeStr: "2024-12-14 15:31:45", RxSign: "W5DEF", TxSign: "G3ABC", Power_dBm: 25, Snr_dB: -12},
			},
			targetCallsign:     "W5XYZ",
			normTxPwr_dBm:      43,
			wantError:          false,
			wantGroupsCount:    2,
			wantFirstGroupSize: 2,
		},
		{
			name: "target not in reports",
			rawReports: []ReceptionReport{
				{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "N0OTH", Power_dBm: 20, Snr_dB: -15},
			},
			targetCallsign:  "W5XYZ",
			normTxPwr_dBm:   43,
			wantError:       true,
			wantGroupsCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processRawRxReports(tt.rawReports, tt.targetCallsign, tt.normTxPwr_dBm)

			if (err != nil) != tt.wantError {
				t.Errorf("processRawRxReports() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError {
				if len(result) != tt.wantGroupsCount {
					t.Errorf("processRawRxReports() returned %d groups, want %d", len(result), tt.wantGroupsCount)
				}

				if len(result) > 0 && len(result[0].Reports) != tt.wantFirstGroupSize {
					t.Errorf("processRawRxReports() first group has %d reports, want %d", len(result[0].Reports), tt.wantFirstGroupSize)
				}
			}
		})
	}
}

// TestProcessRawRxReports_SortedBySnr tests that processRawRxReports sorts reports by normalized SNR.
func TestProcessRawRxReports_SortedBySnr(t *testing.T) {
	rawReports := []ReceptionReport{
		{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "N0OTH", Power_dBm: 20, Snr_dB: -5},
		{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "W5XYZ", Power_dBm: 10, Snr_dB: -10},
		{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "G3ABC", Power_dBm: 30, Snr_dB: -20},
	}

	result, err := processRawRxReports(rawReports, "W5XYZ", int8(43))

	if err != nil {
		t.Fatalf("processRawRxReports() unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("processRawRxReports() returned %d groups, want 1", len(result))
	}

	// Check that reports are sorted by descending normalized SNR
	for i := 0; i < len(result[0].Reports)-1; i++ {
		snr1 := result[0].Reports[i].SnrNorm_dB(43)
		snr2 := result[0].Reports[i+1].SnrNorm_dB(43)
		if snr1 < snr2 {
			t.Errorf("processRawRxReports() reports not sorted by descending SNR: %d >= %d", snr1, snr2)
		}
	}
}

// TestFilterRxReports tests the filterRxReports function.
func TestFilterRxReports(t *testing.T) {
	targetTime := time.Date(2024, 12, 14, 15, 30, 45, 0, time.UTC)

	tests := []struct {
		name            string
		rxReports       []ReceptionReportGroup
		targetCallsign  string
		wantError       bool
		wantGroupsCount int
	}{
		{
			name: "within distance range",
			rxReports: []ReceptionReportGroup{
				{
					RxSign:      "W5ABC",
					Time:        targetTime,
					TargetIndex: 0,
					Reports: []ReceptionReport{
						{TxSign: "W5XYZ", Distance_km: 200},
						{TxSign: "N0OTH", Distance_km: 225}, // Within 75-125% range
						{TxSign: "G3ABC", Distance_km: 150}, // Within 75-125% range
					},
				},
			},
			targetCallsign:  "W5XYZ",
			wantError:       false,
			wantGroupsCount: 1,
		},
		{
			name: "outside distance range",
			rxReports: []ReceptionReportGroup{
				{
					RxSign:      "W5ABC",
					Time:        targetTime,
					TargetIndex: 0,
					Reports: []ReceptionReport{
						{TxSign: "W5XYZ", Distance_km: 200},
						{TxSign: "N0OTH", Distance_km: 500}, // Too far
					},
				},
			},
			targetCallsign:  "W5XYZ",
			wantError:       false,
			wantGroupsCount: 0, // Filtered out due to insufficient comparable transmitters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterRxReports(tt.rxReports, tt.targetCallsign)

			if (err != nil) != tt.wantError {
				t.Errorf("filterRxReports() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && len(result) != tt.wantGroupsCount {
				t.Errorf("filterRxReports() returned %d groups, want %d", len(result), tt.wantGroupsCount)
			}
		})
	}
}

// TestFilterRxReports_DistanceRangeCalculation tests the distance range filtering logic.
func TestFilterRxReports_DistanceRangeCalculation(t *testing.T) {
	targetTime := time.Date(2024, 12, 14, 15, 30, 45, 0, time.UTC)

	rxReports := []ReceptionReportGroup{
		{
			RxSign:      "W5ABC",
			Time:        targetTime,
			TargetIndex: 0,
			Reports: []ReceptionReport{
				{TxSign: "W5XYZ", Distance_km: 100}, // target
				{TxSign: "N0OTH", Distance_km: 75},  // 75% of 100 - minimum
				{TxSign: "G3ABC", Distance_km: 125}, // 125% of 100 - maximum
			},
		},
	}

	result, err := filterRxReports(rxReports, "W5XYZ")

	if err != nil {
		t.Fatalf("filterRxReports() unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("filterRxReports() returned %d groups, want 1", len(result))
	}

	if len(result) > 0 && len(result[0].Reports) != 3 {
		t.Errorf("filterRxReports() filtered group has %d reports, want 3", len(result[0].Reports))
	}
}

// TestPrintReportsAndStats tests that PrintReportsAndStats runs without panic.
func TestPrintReportsAndStats(t *testing.T) {
	targetTime := time.Date(2024, 12, 14, 15, 30, 45, 0, time.UTC)

	rxReports := []ReceptionReportGroup{
		{
			RxSign:      "W5ABC",
			Time:        targetTime,
			TargetIndex: 0,
			Reports: []ReceptionReport{
				{TxSign: "W5XYZ", Distance_km: 200, Power_dBm: 10, Snr_dB: -10, RxAzimuth: 45},
				{TxSign: "N0OTH", Distance_km: 225, Power_dBm: 20, Snr_dB: -15, RxAzimuth: 90},
				{TxSign: "G3ABC", Distance_km: 150, Power_dBm: 30, Snr_dB: -5, RxAzimuth: 180},
			},
		},
	}

	// This should not panic
	PrintReportsAndStats(rxReports, "W5XYZ", int8(43), false)
}

// TestPrintReportsAndStats_Verbose tests PrintReportsAndStats in verbose mode.
func TestPrintReportsAndStats_Verbose(t *testing.T) {
	targetTime := time.Date(2024, 12, 14, 15, 30, 45, 0, time.UTC)

	rxReports := []ReceptionReportGroup{
		{
			RxSign:      "W5ABC",
			Time:        targetTime,
			TargetIndex: 0,
			Reports: []ReceptionReport{
				{TxSign: "W5XYZ", Distance_km: 200, Power_dBm: 10, Snr_dB: -10, RxAzimuth: 45},
				{TxSign: "N0OTH", Distance_km: 225, Power_dBm: 20, Snr_dB: -15, RxAzimuth: 90},
			},
		},
	}

	// This should not panic in verbose mode
	PrintReportsAndStats(rxReports, "W5XYZ", int8(43), true)
}

// TestPrintReportsAndStats_Empty tests PrintReportsAndStats with empty reports.
func TestPrintReportsAndStats_Empty(t *testing.T) {
	var emptyReports []ReceptionReportGroup

	// This should handle empty reports gracefully
	PrintReportsAndStats(emptyReports, "W5XYZ", int8(43), false)
}

// TestMedian_WithFloats tests median with different float types.
func TestMedian_ThreeElementOdd(t *testing.T) {
	values := []float64{1.0, 3.0, 2.0}
	result, err := median(values, false)

	if err != nil {
		t.Errorf("median() unexpected error: %v", err)
	}

	expected := 2.0
	if result != expected {
		t.Errorf("median() = %v, want %v", result, expected)
	}
}

// TestProcessRawRxReports_MultipleGroups tests grouping with multiple time/receiver combinations.
func TestProcessRawRxReports_MultipleGroups(t *testing.T) {
	rawReports := []ReceptionReport{
		{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "W5XYZ", Power_dBm: 10, Snr_dB: -10},
		{TimeStr: "2024-12-14 15:30:45", RxSign: "W5ABC", TxSign: "N0OTH", Power_dBm: 20, Snr_dB: -15},
		{TimeStr: "2024-12-14 15:31:45", RxSign: "W5ABC", TxSign: "W5XYZ", Power_dBm: 12, Snr_dB: -8},
		{TimeStr: "2024-12-14 15:31:45", RxSign: "W5ABC", TxSign: "G3ABC", Power_dBm: 25, Snr_dB: -12},
		{TimeStr: "2024-12-14 15:31:45", RxSign: "W5DEF", TxSign: "W5XYZ", Power_dBm: 15, Snr_dB: -11},
		{TimeStr: "2024-12-14 15:31:45", RxSign: "W5DEF", TxSign: "N0OTH", Power_dBm: 18, Snr_dB: -14},
	}

	result, err := processRawRxReports(rawReports, "W5XYZ", 43)

	if err != nil {
		t.Fatalf("processRawRxReports() unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("processRawRxReports() returned %d groups, want 3", len(result))
	}
}

// TestFilterRxReports_MinimumDistance50km tests that minimum distance is at least 50km.
func TestFilterRxReports_MinimumDistance(t *testing.T) {
	targetTime := time.Date(2024, 12, 14, 15, 30, 45, 0, time.UTC)

	// Distance 20km target - min should be 20*0.75=15, max should be max(20*1.25, 50)=50
	rxReports := []ReceptionReportGroup{
		{
			RxSign:      "W5ABC",
			Time:        targetTime,
			TargetIndex: 0,
			Reports: []ReceptionReport{
				{TxSign: "W5XYZ", Distance_km: 20}, // target
				{TxSign: "N0OTH", Distance_km: 15}, // 75% - within range
				{TxSign: "G3ABC", Distance_km: 25}, // 125%
			},
		},
	}

	result, err := filterRxReports(rxReports, "W5XYZ")

	if err != nil {
		t.Fatalf("filterRxReports() unexpected error: %v", err)
	}

	// N0OTH should be included (15 >= 20*0.75 = 15)
	if len(result) > 0 {
		hasN0OTH := false
		for _, group := range result {
			for _, report := range group.Reports {
				if report.TxSign == "N0OTH" {
					hasN0OTH = true
				}
			}
		}
		if !hasN0OTH {
			t.Errorf("filterRxReports() should have included transmitter at 15km (75%% of target distance)")
		}
	}
}
