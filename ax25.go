package aprsgo

import "fmt"

// ax25.go contains routines for producing a valid ax25 data packet

const (
	flag        byte = 0x7e // the flag byte for each frame is not bit stuffed
	controlFlag byte = 0x03 // Unnumbered Information (UI) frames
	pID         byte = 0xF0 // Protocol IDentifier (PID), no Layer 3 protocol
)

// AX25Data holds the data in an AX25 frame without flags
type AX25Data []byte

// SSID for station identification or path indication
type SSID byte

// PositionData contains the data to construct an APRS position report
type PositionData struct {
	Callsign        string // limited to 6 ASCII characters
	StationSSID     SSID
	Destination     string // limited to 6 ASCII characters
	DestinationSSID SSID
	Latitude        float64
	Longitude       float64
	Altitude        float64
	Course          float64
	Speed           float64
	Comment         string
}

// Destination SSID codes for AX.25 destination address fields
const (
	DestSSIDVIAPath SSID = iota
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

// BasicAPRSReport constructs a basic APRS position report
func (data PositionData) BasicAPRSReport() AX25Data {

	destinationAddress := constructAddress(data.Destination, data.DestinationSSID)
	sourceAddress := constructAddress(data.Callsign, data.StationSSID)
	informationField := CalculateBasicInformationField(data)

	ax25data := AssembleAX25Data(sourceAddress, destinationAddress, informationField)

	return AX25Data(ax25data)
}

// CalculateBasicInformationField for an APRS position report
func CalculateBasicInformationField(report PositionData) []byte {
	// the information field containing lat/long in text format
	dataTypeIdentifier := "!" // realtime position with no messaging

	latDir := "N" // default 0 is N
	lat := report.Latitude
	if lat < 0 {
		latDir = "S"
		lat = -lat
	}
	latDeg := int(lat) // TODO to check for proper bounds
	latMin := 60.0 * (lat - float64(latDeg))

	longDir := "W" // default 0 is W
	long := report.Longitude
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
		report.Comment)

	return []byte(informationField)
}

func constructAddress(callsign string, ssid SSID) [7]byte {
	var address [7]byte
	copy(address[:], "      ") // intialize the field with spaces for shorter callsigns
	copy(address[:], callsign) // truncate any longer callsigns
	address[6] = byte(ssid)    // write the SSID into the 7-th byte
	return address
}

// AssembleAX25Data converts raw address, destination and information fields into an unnumbered information AX25 UI packet
func AssembleAX25Data(sourceAddress [7]byte, destinationAddress [7]byte, informationField []byte) []byte {
	var ax25data []byte

	// append the addresses, destination first
	for i := 0; i < 7; i++ {
		destinationAddress[i] = destinationAddress[i] << 1 // AX.25 addresses are shifted left one bit
	}
	ax25data = append(ax25data, destinationAddress[:]...)
	for i := 0; i < 7; i++ {
		sourceAddress[i] = sourceAddress[i] << 1 // AX.25 addresses are shifted left one bit
	}

	// note that AX25 data allows additional address fields here as well, not currently supported

	sourceAddress[6]++ // the final address bit is set to one
	ax25data = append(ax25data, sourceAddress[:]...)

	// standard flags for APRS
	ax25data = append(ax25data, controlFlag)
	ax25data = append(ax25data, pID)

	// the information fields
	ax25data = append(ax25data, informationField...)

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
