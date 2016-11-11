package aprsgo

import (
	"testing"
)

func TestNewBasicAPRSAX25Data(t *testing.T) {
	report := PositionData{
		Callsign:  "W1AW",
		Latitude:  41.7147,
		Longitude: -72.7272,
		Comment:   "Test",
	}

	ax25data := report.BasicAPRSReport()
	// decode the address fields
	for i := 0; i < 14; i++ {
		ax25data[i] = ax25data[i] >> 1
	}
	t.Log(string(ax25data))
}
