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

func TestBase91Encode(t *testing.T) {
	testCases := []struct {
		In   uint32
		Want string
	}{
		{In: uint32(12345678),
			Want: "1Cmi",
		},
		{In: uint32(20427156),
			Want: "<*e7",
		},
	}

	for _, testCase := range testCases {
		got, err := Base91Encode(testCase.In, 4)
		if err != nil {
			t.Errorf("Encoding %d failed with error %v", testCase.In, err)
		}
		if got != testCase.Want {
			t.Errorf("Encoding %d, expected: %s got: %s", testCase.In, testCase.Want, got)
		}
		_, err = Base91Encode(testCase.In, 3) // too few digits to encode the number
		if err == nil {
			t.Errorf("Encoding %d expected to fail when using 3 digits, but didn't", testCase.In)
		} else {
			t.Logf("Encoding %d expected to fail when using 3 digits, %v", testCase.In, err)
		}
	}
}

func TestNewCompressedAPRSAX25Data(t *testing.T) {
	report := PositionData{
		Callsign:  "W1AW",
		Latitude:  41.7147,
		Longitude: -72.7272,
		Comment:   "Test",
	}

	ax25data := report.CompressedAPRSReport()
	// decode the address fields
	for i := 0; i < 14; i++ {
		ax25data[i] = ax25data[i] >> 1
	}
	t.Log(string(ax25data))
}
