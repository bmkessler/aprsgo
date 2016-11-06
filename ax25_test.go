package aprsgo

import (
	"testing"
)

func TestNewBasicAPRSAX25Data(t *testing.T) {
	report := PositionReport{
		Callsign:        "W1AW",
		Destination:     VersionDestinationAddress,
		DestinationSSID: SSID(1),
		Latitude:        41.7147,
		Longitude:       -72.7272,
		Comment:         "Test",
	}

	ax25data := BasicAPRSAX25Data(report)
	// decode the address fields
	for i := 0; i < 14; i++ {
		ax25data[i] = ax25data[i] >> 1
	}
	t.Log(string(ax25data))
}
