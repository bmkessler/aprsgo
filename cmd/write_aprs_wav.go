package main

import "github.com/bmkessler/aprsgo"
import "log"
import "flag"

const version = "APZ001"

func main() {

	var sampleRate = flag.Uint("sr", 48000, "Sample rate in samples per second")
	var bitRate = flag.Uint("br", 16, "Bit rate in bits per sample, 8, 16, 24, and 32 supported")
	var callsign = flag.String("call", "KK6ZNQ", "Callsign to send from")
	var comment = flag.String("comment", "Test", "Comment to append to position report")
	var lat = flag.Float64("lat", 37.7772103, "Latitude for position report")
	var long = flag.Float64("long", -122.4499289, "Longitude for position report")
	var digipath = flag.Uint("digi", 0, "The SSID to indicate the digipeater path")
	var filename = flag.String("file", "test_file.wav", "The output filename")

	flag.Parse()

	wr, err := aprsgo.NewWaveWriter(uint32(*sampleRate), uint8(*bitRate), 1)
	if err != nil {
		log.Fatal(err)
	}
	ax25data := aprsgo.NewBasicAPRSAX25Data(*callsign, *lat, *long, *comment, version, byte(*digipath))
	aprsgo.WriteSamples(ax25data, wr)
	wr.WriteFile(*filename)
}

// To test the output file with multimon the WAV file can be piped with sox to the expected format
// sox -t wav test_file.wav -esigned-integer -b16 -r 22050 -t raw - | multimon-ng -a AFSK1200 -A -t raw -
