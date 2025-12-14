package wspranalysis

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"golang.org/x/exp/constraints"
)

// Build a new instance of ReceptionReportGroup (see types.go) from a slice of
// ReceptionReports. Return an error if targetCallsign is not found in reports.
func newReceptionReportGroup(reports []ReceptionReport, targetCallsign string) (*ReceptionReportGroup, error) {
	newGroup := new(ReceptionReportGroup)
	if len(reports) > 0 {
		newGroup.RxSign = reports[0].RxSign
		newGroup.Time = reports[0].Time()
		newGroup.Reports = reports
		newGroup.TargetIndex = -1
		for i, report := range newGroup.Reports {
			if report.TxSign == targetCallsign {
				newGroup.TargetIndex = i
				break
			}
		}
		if newGroup.TargetIndex == -1 {
			return nil, fmt.Errorf("target transmitter %s not found in report group for receiver %s at time %s",
				targetCallsign, newGroup.RxSign,
				newGroup.Time.UTC().Format(time.RFC3339))
		}
	}
	return newGroup, nil
}

// Helper function to calculate the median of an optionally pre-sorted slice.
// NOTE: If preSorted is false, the input slice will be sorted in place.
func median[T interface {
	constraints.Integer | constraints.Float
}](values []T, preSorted bool) (float64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("cannot calculate median of empty slice")
	}
	if !preSorted {
		slices.Sort(values)
	}
	if len(values) == 1 {
		return float64(values[0]), nil
	} else if len(values)%2 == 0 {
		midIndexA := len(values)/2 - 1
		midIndexB := len(values) / 2
		return (float64(values[midIndexA]) + float64(values[midIndexB])) / 2, nil
	} else {
		midIndex := len(values) / 2
		return float64(values[midIndex]), nil
	}
}

// This function groups the raw reports returned by the database query into chunks associated
// with a particular receiver and time. Within each chunk, the reports are ordered by
// descending normalised SNR. The normalised SNR is based on a notional transmit power of
// normTxPower_dBm.
// The function returns a slice of ReceptionReportGroup structs, with each entry containing
// the reports for a particular receiver and time. The slice is ordered by time followed by
// receiver callsign.
func processRawRxReports(rawRxReports []ReceptionReport, targetCallsign string, normTxPwr_dBm int8) ([]ReceptionReportGroup, error) {
	// rawRxReports is one-dimensional and is ordered by time, followed by receiver callsign.
	// We need to split it each time the time or receiver field changes and build a
	// ReceptionReportGroup struct.
	var rxReports []ReceptionReportGroup
	for i, j := 0, 1; j <= len(rawRxReports); j++ {
		if j == len(rawRxReports) ||
			rawRxReports[j].TimeStr != rawRxReports[i].TimeStr ||
			rawRxReports[j].RxSign != rawRxReports[i].RxSign {
			// The slice rawRxReports[i:j] forms a report group.
			reportsForGroup := rawRxReports[i:j]
			// Sort it by descending normalised SNR.
			slices.SortFunc(reportsForGroup, func(a, b ReceptionReport) int {
				return cmp.Compare(b.SnrNorm_dB(normTxPwr_dBm), a.SnrNorm_dB(normTxPwr_dBm))
			})
			// Build a ReceptionReportGroup struct and append it to rxReports.
			newGroup, err := newReceptionReportGroup(reportsForGroup, targetCallsign)
			if err != nil {
				return nil, err
			}
			if newGroup == nil || len(newGroup.Reports) == 0 {
				return nil, fmt.Errorf("generated nil/empty report group. This should not happen")
			}
			rxReports = append(rxReports, *newGroup)
			// Move to the next group.
			i = j
		}
	}
	return rxReports, nil
}

// Filter reception reports to remove transmitters which are not comparable to
// the target transmitter. Currently this is just based on distance from the
// receiver.
func filterRxReports(rxReports []ReceptionReportGroup, targetCallsign string) ([]ReceptionReportGroup, error) {
	var filteredReports []ReceptionReportGroup
	for _, reportGroup := range rxReports {
		// Find the distance of the target transmitter from the receiver in order to
		// establish upper and lower bounds on distance for comparable transmitters.
		targetReport := reportGroup.Reports[reportGroup.TargetIndex]
		// Define acceptable distance range as +/- 25% of target distance.
		var distanceMin_km uint16 = uint16(float64(targetReport.Distance_km) * 0.75)
		var distanceMax_km uint16 = max(uint16(float64(targetReport.Distance_km)*1.25), 50)
		// Build a new report group containing only the reports within the acceptable distance range.
		filteredListForGroup := make([]ReceptionReport, 0, len(reportGroup.Reports))
		for _, report := range reportGroup.Reports {
			if report.Distance_km >= distanceMin_km && report.Distance_km <= distanceMax_km {
				filteredListForGroup = append(filteredListForGroup, report)
			}
		}
		newReportGroup, err := newReceptionReportGroup(filteredListForGroup, targetCallsign)
		if err != nil {
			return nil, fmt.Errorf("error building filtered report group (%w)", err)
		}
		if newReportGroup == nil || len(newReportGroup.Reports) < 2 {
			// Skip groups with insufficient comparable transmitters.
			fmt.Printf("Reports from %s at %s filtered out due to insufficient comparable transmitters\n", reportGroup.RxSign, reportGroup.Time.UTC().Format(time.RFC3339))
		} else {
			filteredReports = append(filteredReports, *newReportGroup)
		}
	}
	return filteredReports, nil
}

// Print out the reception reports nicely formatted to the console. Also perform some
// basic stats to show how the target transmitter compares with the rest.
func PrintReportsAndStats(rxReports []ReceptionReportGroup, targetCallsign string, normTxPwr_dBm int8, verbose bool) {
	var aggregatedRelativeSnrNorms []int8
	for _, reportGroup := range rxReports {
		fmt.Printf("Received by %s (distance %dkm) at %s:\n", reportGroup.RxSign, reportGroup.Reports[reportGroup.TargetIndex].Distance_km,
			reportGroup.Time.UTC().Format("2006-01-02T15:04:05Z07:00"))
		for i, report := range reportGroup.Reports {
			if verbose {
				if i == reportGroup.TargetIndex {
					fmt.Printf("     -->")
				} else {
					fmt.Printf("        ")
				}
				fmt.Printf("%d: Transmitter: %s, Power: %ddBm, Distance: %dkm, RX Azimuth: %dº, SNR: %+ddB, Normalised SNR: %+ddB\n", i+1,
					report.TxSign, report.Power_dBm, report.Distance_km, report.RxAzimuth, report.Snr_dB, report.SnrNorm_dB(normTxPwr_dBm))
			}
			if i != reportGroup.TargetIndex {
				relativeSnrNorm := report.SnrNorm_dB(normTxPwr_dBm) - reportGroup.Reports[reportGroup.TargetIndex].SnrNorm_dB(normTxPwr_dBm)
				aggregatedRelativeSnrNorms = append(aggregatedRelativeSnrNorms, relativeSnrNorm)
			}
		}
		if len(reportGroup.Reports) > 1 {
			var medianSnrNorm_dB int8
			reports := reportGroup.Reports
			if len(reportGroup.Reports)%2 == 0 {
				medianSnrNorm_dB = (reports[len(reports)/2-1].SnrNorm_dB(normTxPwr_dBm) + reports[len(reports)/2].SnrNorm_dB(normTxPwr_dBm)) / 2
			} else {
				medianSnrNorm_dB = reports[len(reports)/2].SnrNorm_dB(normTxPwr_dBm)
			}
			targetSnrNorm := reports[reportGroup.TargetIndex].SnrNorm_dB(normTxPwr_dBm)
			targetSnrNorm_dBmedian := targetSnrNorm - medianSnrNorm_dB
			fmt.Printf("    %d out of %d transmitters; Normalised SNR: %+ddB, %+ddBmedian\n", reportGroup.TargetIndex+1, len(reports), targetSnrNorm, targetSnrNorm_dBmedian)
		}
	}
	fmt.Printf("\nOffset from median of relative normalised SNR of all other transmitters: ")
	if l := len(aggregatedRelativeSnrNorms); l > 1 {
		aggregatedMedian, _ := median(aggregatedRelativeSnrNorms, false)
		fmt.Printf("%+.1fdBmedian (%d samples)\n", -aggregatedMedian, l)
	}
}

// RunAnalysis orchestrates the query, filtering and printing. This is function
// called by main.go.
func RunAnalysis(targetCallsign string, band int, startTime time.Time, duration time.Duration, normTxPwr_dBm int8, verbose bool) error {
	// Run the database query to get raw reception reports.
	rawRxReports, err := RunQuery[ReceptionReport](BuildQueryUrl(targetCallsign, band, startTime, duration))
	if err != nil {
		return fmt.Errorf("error running database query on wspr.live (%w)", err)
	}
	if len(rawRxReports) == 0 {
		return fmt.Errorf("no reception reports found for %s on band %d in the specified time range", targetCallsign, band)
	}
	// Process the raw reception reports into structured groups.
	rxReports, err := processRawRxReports(rawRxReports, targetCallsign, normTxPwr_dBm)
	if err != nil {
		return err
	}
	// Filter the reception reports to remove non-comparable transmitters.
	rxReports, err = filterRxReports(rxReports, targetCallsign)
	if err != nil {
		return err
	}
	// Print out the reports and stats.
	PrintReportsAndStats(rxReports, targetCallsign, normTxPwr_dBm, verbose)
	return nil
}
