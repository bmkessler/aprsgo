package aprsgo

// ax25.go contains routines for producing a valid ax25 data packet

const (
	flag        byte = 0x7e // the flag byte for each frame is not bit stuffed
	controlFlag byte = 0x03 // Unnumbered Information (UI) frames
	pID         byte = 0xF0 // Protocol IDentifier (PID), no Layer 3 protocol
)

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
