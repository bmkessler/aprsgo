package aprsgo

import "fmt"

// ax25.go contains routines for producing a valid ax25 data packet

const (
	flag        byte = 0x7e // the flag byte for each frame is not bit stuffed
	controlFlag byte = 0x03 // Unnumbered Information (UI) frames
	pID         byte = 0xF0 // Protocol IDentifier (PID), no Layer 3 protocol
)

// Destination SSID codes for AX.25 destination address fields
const (
	DestSSIDVIAPath byte = iota
	DestSSIDWIDE1_1Path
	DestSSIDWIDE2_2Path
	DestSSIDWIDE3_3Path
	DestSSIDWIDE4_4Path
	DestSSIDWIDE5_5Path
	DestSSIDWIDE6_6Path
	DestSSIDWIDE7_7Path
	DestSSIDNorthPath
	DestSSIDSouthPath
	DestSSIDEastPath
	DestSSIDWestPath
	DestSSIDNorthPathWIDE
	DestSSIDSouthPathWIDE
	DestSSIDEastPathWIDE
	DestSSIDWestPathWIDE
)

// VersionDestinationAddress is the address designating the software version
var VersionDestinationAddress = "APZ001"

// NewBasicAPRSAX25Data calculates an ax.25 data frame in the APRS format for text lat/long
func NewBasicAPRSAX25Data(callsign string, lat, long float64, comment string, destination string, destinationSSID byte) []byte {
	var ax25data []byte

	// the address fields destination first
	var destinationBytes [6]byte
	copy(destinationBytes[:], "      ")    // intialize the field with spaces
	copy(destinationBytes[:], destination) // TODO to check for proper bounds
	for i := 0; i < 6; i++ {
		destinationBytes[i] = destinationBytes[i] << 1 // AX.25 addresses are shifted left one bit
	}
	ax25data = append(ax25data, destinationBytes[:]...)

	destinationSSID = destinationSSID << 1 // the digipeater path
	ax25data = append(ax25data, destinationSSID)

	var callsignBytes [6]byte
	copy(callsignBytes[:], "      ")
	copy(callsignBytes[:], callsign) // TODO to check for proper bounds
	for i := 0; i < 6; i++ {
		callsignBytes[i] = callsignBytes[i] << 1 // AX.25 addresses are shifted left one bit
	}
	ax25data = append(ax25data, callsignBytes[:]...)

	callsignSSID := byte(0)                // use the symbol from the information field
	callsignSSID = (callsignSSID << 1) + 1 // final address byte has 1 in lowest bit
	ax25data = append(ax25data, callsignSSID)

	// standard flags for APRS
	ax25data = append(ax25data, controlFlag)
	ax25data = append(ax25data, pID)

	// the information field containing lat/long in text format
	dataTypeIdentifier := "!" // realtime position with no messaging

	latDir := "N" // default 0 is N
	if lat < 0 {
		latDir = "S"
		lat = -lat
	}
	latDeg := int(lat) // TODO to check for proper bounds
	latMin := 60.0 * (lat - float64(latDeg))

	longDir := "W" // default 0 is W
	if long > 0 {
		longDir = "E"
	} else {
		long = -long
	}
	longDeg := int(long) // TODO check for proper bounds
	longMin := 60.0 * (long - float64(longDeg))

	displaySymbolTableIdentifier := "/" // primary table
	displaySymbol := "-"                // house

	informationField := fmt.Sprintf("%s%02d%2.2f%s%s%3d%2.2f%s%s%s",
		dataTypeIdentifier,
		latDeg,
		latMin,
		latDir,
		displaySymbolTableIdentifier,
		longDeg,
		longMin,
		longDir,
		displaySymbol,
		comment)

	ax25data = append(ax25data, []byte(informationField)...)

	ax25data = appendFCS(ax25data) // append the 16-bit CRC Frame Check Sequence
	return ax25data
}

func calcCRC(ax25data []byte) uint16 {
	// calculates the CRC-16-CCITT for an array of bytes
	var crc uint16 = 0xffff
	for _, axByte := range ax25data {
		for i := 0; i < 8; i++ {
			bit := axByte & 0x0001
			if (crc & 0x0001) != uint16(bit) {
				crc = (crc >> 1) ^ 0x8408
			} else {
				crc = crc >> 1
			}
			axByte = axByte >> 1
		}
	}
	return crc ^ 0xffff
}

func appendFCS(ax25data []byte) []byte {
	fcs := calcCRC(ax25data)            // calculate the CRC value
	fcsMSB := byte(fcs & 0x00FF)        // the most significant byte
	fcsLSB := byte((fcs >> 8) & 0x00FF) // the least significant byte

	ax25data = append(ax25data, fcsMSB)
	ax25data = append(ax25data, fcsLSB)
	return ax25data
}
