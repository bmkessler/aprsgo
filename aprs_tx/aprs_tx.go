package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bmkessler/aprsgo"
)

func main() {
	// position report parameters
	callsign := flag.String("call", "W1AW", "Callsign to send from")
	comment := flag.String("comment", "Test", "Comment to append to position report")
	lat := flag.Float64("lat", 41.7147, "Latitude for position report")
	long := flag.Float64("long", -72.7272, "Longitude for position report")
	format := flag.String("format", "b", "Format for position report, 'b'=basic, 'c'=compressed")
	// WAV file parameters
	sampleRate := flag.Uint("sr", 48000, "Sample rate in samples per second")
	bitRate := flag.Uint("br", 16, "Bit rate in bits per sample, 8, 16, 24, and 32 supported")
	numChannels := flag.Uint("nc", 1, "Number of audio channels to record")

	flag.Parse()

	report := aprsgo.PositionData{
		Callsign:  *callsign,
		Latitude:  *lat,
		Longitude: *long,
		Comment:   *comment,
	}

	var ax25data aprsgo.AX25Data
	switch *format {
	case "c": // compressed
		ax25data = report.CompressedAPRSReport()
	default: // "b" and anything else not recognized
		ax25data = report.BasicAPRSReport()
	}

	symbolStream := ax25data.Encode()

	wavFilename := fmt.Sprintf("%s_%.2f_%.2f_%dHz_%dbits_%dchan_%s.wav",
		*callsign,
		*lat,
		*long,
		*sampleRate,
		*bitRate,
		*numChannels,
		*comment)

	params := aprsgo.WAVParams{
		Filename:         wavFilename,
		SamplesPerSecond: uint32(*sampleRate),
		BitsPerSample:    uint8(*bitRate),
		NumChannels:      uint8(*numChannels),
	}

	if err := symbolStream.WriteWAV(params); err != nil {
		log.Fatal(err)
	}
}

// To test the output file with multimon-ng the WAV file can be piped with sox to the expected format
// sox -t wav test_file.wav -esigned-integer -b16 -r 22050 -t raw - | multimon-ng -a AFSK1200 -A -t raw -
//
// expected output should be:
// APRS: W1AW>APZ001:!4142.88N/ 7243.63W-Test
