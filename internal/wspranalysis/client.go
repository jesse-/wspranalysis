// This file contains the code for interacting with the wspr.live database.
package wspranalysis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Base URL for querying the wspr.live database.
const baseQueryURL string = "https://db1.wspr.live/?query="

// Build a query URL to ask wspr.live for all the reception reports of the
// target transmitter within the specified time range. Additionally, list all
// the other transmitters which were received alongside the target transmitter.
//
//	txSign: Callsign of the target transmitter.
//	band: Integer code of the band (see bandNameToCode in types.go and
//	      https://wspr.live/ under 'Bands Table').
//	tStart: Start time for the query.
//	duration: Query for reception reports up to duration after tStart.
func BuildQueryUrl(txSign string, band int, tStart time.Time, duration time.Duration) string {
	// The outer SQL query just selects the desired columns for the specified
	// band and time range (this will include all transmitters and receivers).
	query := fmt.Sprintf("SELECT tx_sign, rx_sign, time, power, distance, rx_azimuth, snr FROM wspr.rx AS R WHERE "+
		"band = %d AND "+
		"time >= '%s' AND "+
		"time < '%s' AND "+
		// This nested EXISTS query filters the results with the condition that the
		// same receiver must also have received the target transmitter at the same
		// time and on the same band.
		"EXISTS (SELECT 1 FROM wspr.rx AS S WHERE S.tx_sign = '%s' AND S.band = %d AND S.rx_sign = R.rx_sign AND S.time = R.time) "+
		"ORDER BY time ASC, rx_sign ASC FORMAT JSON",
		band, tStart.UTC().Format(time.DateTime), tStart.UTC().Add(duration).Format(time.DateTime), txSign, band)
	return baseQueryURL + url.PathEscape(query)
}

// Perform the actual HTTP GET request to queryURL and unmarshal the JSON
// response into a slice of ListElementStructs.
func RunQuery[ListElementStruct any](queryURL string) ([]ListElementStruct, error) {
	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("http.Get() failed (%w)", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body (%w)", err)
	}
	var receiverStruct struct {
		Data []ListElementStruct `json:"data"`
	}
	if err := json.Unmarshal(body, &receiverStruct); err != nil {
		return nil, fmt.Errorf("failed to decode JSON (%w)", err)
	}
	return receiverStruct.Data, nil
}
