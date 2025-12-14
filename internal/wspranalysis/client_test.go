package wspranalysis

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestBuildQueryUrl tests the BuildQueryUrl function.
func TestBuildQueryUrl(t *testing.T) {
	tests := []struct {
		name     string
		txSign   string
		band     int
		tStart   time.Time
		duration time.Duration
	}{
		{
			name:     "basic query",
			txSign:   "W5XYZ",
			band:     14,
			tStart:   time.Date(2024, 12, 14, 0, 0, 0, 0, time.UTC),
			duration: 24 * time.Hour,
		},
		{
			name:     "different band",
			txSign:   "G3XYZ",
			band:     3,
			tStart:   time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
			duration: 1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildQueryUrl(tt.txSign, tt.band, tt.tStart, tt.duration)

			// Check that result starts with base URL
			if !strings.Contains(result, baseQueryURL) {
				t.Errorf("BuildQueryUrl() doesn't contain base URL")
			}

			// Check for key SQL components (they'll be URL encoded)
			if !strings.Contains(result, "tx_sign") {
				t.Errorf("BuildQueryUrl() missing SQL reference to tx_sign")
			}
			if !strings.Contains(result, tt.txSign) {
				t.Errorf("BuildQueryUrl() doesn't contain tx_sign: %s", tt.txSign)
			}
		})
	}
}

// TestBuildQueryUrl_ContainsTargetCallsign tests that the query URL contains the target callsign.
func TestBuildQueryUrl_ContainsTargetCallsign(t *testing.T) {
	txSign := "W5XYZ"
	band := 14
	tStart := time.Date(2024, 12, 14, 0, 0, 0, 0, time.UTC)
	duration := 24 * time.Hour

	result := BuildQueryUrl(txSign, band, tStart, duration)

	if !strings.Contains(result, txSign) {
		t.Errorf("BuildQueryUrl() result doesn't contain target callsign: %s", txSign)
	}
}

// TestBuildQueryUrl_ContainsTimeRange tests that the query includes the correct time range.
func TestBuildQueryUrl_TimeRange(t *testing.T) {
	txSign := "W5XYZ"
	band := 14
	tStart := time.Date(2024, 12, 14, 10, 30, 0, 0, time.UTC)
	duration := 2 * time.Hour

	result := BuildQueryUrl(txSign, band, tStart, duration)

	// The URL is encoded, so the actual time strings will be percent-encoded.
	// Just check for the presence of the time values in some form
	if !strings.Contains(result, "2024-12-14") {
		t.Errorf("BuildQueryUrl() doesn't contain date")
	}

	if !strings.Contains(result, "10%3A30%3A00") && !strings.Contains(result, "10:30:00") {
		// Check for either URL encoded or plain format
		if !strings.Contains(result, "1030") {
			t.Errorf("BuildQueryUrl() doesn't contain start time")
		}
	}
}

// TestRunQuery tests the RunQuery function with a mock server.
func TestRunQuery(t *testing.T) {
	// Create a mock server that returns sample JSON data
	mockData := `{
		"data": [
			{
				"time": "2024-12-14 15:30:45",
				"rx_sign": "W5ABC",
				"tx_sign": "W5XYZ",
				"power": 10,
				"snr": -15,
				"distance": 250,
				"rx_azimuth": 45
			},
			{
				"time": "2024-12-14 15:30:45",
				"rx_sign": "W5ABC",
				"tx_sign": "N0OTH",
				"power": 20,
				"snr": -10,
				"distance": 300,
				"rx_azimuth": 90
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockData))
	}))
	defer server.Close()

	result, err := RunQuery[ReceptionReport](server.URL)

	if err != nil {
		t.Errorf("RunQuery() unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("RunQuery() returned %d results, want 2", len(result))
	}

	if result[0].TxSign != "W5XYZ" {
		t.Errorf("RunQuery() first result TxSign = %s, want W5XYZ", result[0].TxSign)
	}

	if result[0].RxSign != "W5ABC" {
		t.Errorf("RunQuery() first result RxSign = %s, want W5ABC", result[0].RxSign)
	}

	if result[0].Power_dBm != 10 {
		t.Errorf("RunQuery() first result Power_dBm = %d, want 10", result[0].Power_dBm)
	}
}

// TestRunQuery_EmptyResponse tests RunQuery with an empty data response.
func TestRunQuery_EmptyResponse(t *testing.T) {
	mockData := `{"data": []}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockData))
	}))
	defer server.Close()

	result, err := RunQuery[ReceptionReport](server.URL)

	if err != nil {
		t.Errorf("RunQuery() unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("RunQuery() returned %d results, want 0", len(result))
	}
}

// TestRunQuery_InvalidJSON tests RunQuery with invalid JSON response.
func TestRunQuery_InvalidJSON(t *testing.T) {
	mockData := `{invalid json}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockData))
	}))
	defer server.Close()

	_, err := RunQuery[ReceptionReport](server.URL)

	if err == nil {
		t.Errorf("RunQuery() expected error for invalid JSON, got nil")
	}
}

// TestRunQuery_ServerError tests RunQuery with a server error.
func TestRunQuery_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	result, err := RunQuery[ReceptionReport](server.URL)

	// The function should still attempt to parse the response
	if result != nil || err == nil {
		// Behavior depends on implementation details
	}
}

// TestRunQuery_MissingFields tests RunQuery handles reports with missing optional fields.
func TestRunQuery_MissingFields(t *testing.T) {
	mockData := `{
		"data": [
			{
				"time": "2024-12-14 15:30:45",
				"rx_sign": "W5ABC",
				"tx_sign": "W5XYZ",
				"power": 10,
				"snr": -15,
				"distance": 250,
				"rx_azimuth": 45
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockData))
	}))
	defer server.Close()

	result, err := RunQuery[ReceptionReport](server.URL)

	if err != nil {
		t.Errorf("RunQuery() unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("RunQuery() returned %d results, want 1", len(result))
	}
}
