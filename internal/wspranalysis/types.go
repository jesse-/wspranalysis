// General definitions and types used throughout the package.
package wspranalysis

import (
	"fmt"
	"time"
)

// Struct to hold a single WSPR reception report. This maps to the JSON
// returned by the wspr.live database query.
type ReceptionReport struct {
	TimeStr     string `json:"time"`
	RxSign      string `json:"rx_sign"`
	TxSign      string `json:"tx_sign"`
	Power_dBm   int8   `json:"power"`
	Snr_dB      int8   `json:"snr"`
	Distance_km uint16 `json:"distance"`
	RxAzimuth   uint16 `json:"rx_azimuth"`
}

// Method to parse the TimeStr field of the above struct into a time.Time
// object.
func (r *ReceptionReport) Time() time.Time {
	t, err := time.Parse(time.DateTime, r.TimeStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

// Method to calculate the normalised SNR for this reception report given a
// reference transmitter power in dBm. The normalised SNR is the SNR that this
// transmission would have if the transmitter were transmitting at txRefPower_dBm.
func (r *ReceptionReport) SnrNorm_dB(txRefPower_dBm int8) int8 {
	return r.Snr_dB + txRefPower_dBm - r.Power_dBm
}

// Struct to hold a group of reception reports received by a single station at
// a given time, including the index of the target transmitter's report within
// the group.
type ReceptionReportGroup struct {
	RxSign      string
	Time        time.Time
	Reports     []ReceptionReport
	TargetIndex int
}

// Map between common band names and their corresponding integer codes used by
// wspr.live.
var bandNameToCode = map[string]int{
	"lf":   -1,
	"mf":   0,
	"160m": 1,
	"80m":  3,
	"60m":  5,
	"40m":  7,
	"30m":  10,
	"20m":  14,
	"17m":  18,
	"15m":  21,
	"12m":  24,
	"10m":  28,
	"6m":   50,
	"4m":   70,
	"2m":   144,
	"70cm": 432,
	"23cm": 1296,
}

// Return all the band names (useful for the CLI help text).
func BandNames() []string {
	names := make([]string, 0, len(bandNameToCode))
	for k := range bandNameToCode {
		names = append(names, k)
	}
	return names
}

// Convert a band name to its corresponding integer code. Returns an error
// if the band name is not recognised.
func BandNameToCode(bandName string) (int, error) {
	if code, ok := bandNameToCode[bandName]; ok {
		return code, nil
	}
	return 0, fmt.Errorf("unrecognised band name: %s", bandName)
}
