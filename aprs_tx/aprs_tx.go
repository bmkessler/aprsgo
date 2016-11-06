package main

import (
	"flag"
	"log"

	"github.com/bmkessler/aprsgo"
)

func main() {
	// position report parameters
	callsign := flag.String("call", "W1AW", "Callsign to send from")
	comment := flag.String("comment", "Test", "Comment to append to position report")
	lat := flag.Float64("lat", 41.7147, "Latitude for position report")
	long := flag.Float64("long", -72.7272, "Longitude for position report")
	digipath := flag.Uint("digi", 0, "The SSID to indicate the digipeater path")
	// WAV file parameters
	filename := flag.String("file", "test_file.wav", "The output filename")
	sampleRate := flag.Uint("sr", 48000, "Sample rate in samples per second")
	bitRate := flag.Uint("br", 16, "Bit rate in bits per sample, 8, 16, 24, and 32 supported")
	numChannels := flag.Uint("nc", 1, "Number of audio channels to record")

	flag.Parse()

	report := aprsgo.PositionReport{
		Callsign:        *callsign,
		Destination:     aprsgo.VersionDestinationAddress,
		DestinationSSID: aprsgo.SSID(*digipath),
		Latitude:        *lat,
		Longitude:       *long,
		Comment:         *comment,
	}

	ax25data := aprsgo.BasicAPRSAX25Data(report)

	symbolStream := aprsgo.EncodeAX25Data(ax25data)

	params := aprsgo.WAVParams{
		Filename:         *filename,
		SamplesPerSecond: uint32(*sampleRate),
		BitsPerSample:    uint8(*bitRate),
		NumChannels:      uint8(*numChannels),
	}

	if err := aprsgo.WriteWAV(symbolStream, params); err != nil {
		log.Fatal(err)
	}
}

// To test the output file with multimon the WAV file can be piped with sox to the expected format
// sox -t wav test_file.wav -esigned-integer -b16 -r 22050 -t raw - | multimon-ng -a AFSK1200 -A -t raw -
//
// expected output should be:
// APRS: W1AW>APZ001:!4142.88N/ 7243.63W-Test
