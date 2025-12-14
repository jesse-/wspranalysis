// Main entry point for the wspranalysis command-line tool.
// The code here mostly just handles argument parsing. The main
// analysis logic is in the internal/wspranalysis package.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jesse-/wspranalysis/internal/wspranalysis"
)

func main() {
	defaultStartTimeStr := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	normTxPwr := flag.Int("norm", 43, "Transmit power in dBm to normalise SNRs for.")
	startTimeStr := flag.String("start", defaultStartTimeStr, "`Start time` for the query in RFC3339 format")
	duration := flag.Duration("duration", 24*time.Hour, "Duration to analyse over (e.g., 24h, 30m)")
	verbose := flag.Bool("v", false, "Enable verbose output")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [target callsign] [band]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Query the wspr.live database for reception reports of [target callsign] on [band].\n")
		fmt.Fprintf(os.Stderr, "Each reception report is ranked against other transmitters heard by the same receiver\n")
		fmt.Fprintf(os.Stderr, "at the same time.\n\n")
		fmt.Fprintf(os.Stderr, "[band] is one of:\n\t%v\n\n", wspranalysis.BandNames())
		fmt.Fprintf(os.Stderr, "Other options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		return
	}
	target := strings.ToUpper(flag.Args()[0])
	bandName := strings.ToLower(flag.Args()[1])
	band, err := wspranalysis.BandNameToCode(bandName)
	if err != nil {
		flag.Usage()
		return
	}
	startTime, err := time.Parse(time.RFC3339, *startTimeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing start time: %v\n", err)
		return
	}
	if *normTxPwr < -128 || *normTxPwr > 127 {
		fmt.Fprintf(os.Stderr, "Error: Normalised transmit power must be between -128 and 127 dBm\n")
		return
	}

	if err := wspranalysis.RunAnalysis(target, band, startTime, *duration, int8(*normTxPwr), *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
