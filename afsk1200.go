package aprsgo

// This file contains the routines for writing out a WAV file
// of the AFSK1200 data with Non-Return to Zero Inverted (NRZI)
// encoding of the data, bit stuffing and appropriate start and end flags

type symbol bool

const (
	flagByte     byte    = 0x7e   // the flag bit for each frame is not bit stuffed
	symbolRate   float64 = 1200   // 1200 baud
	markFreq     float64 = 1200.0 // mark frequency in Hertz
	spaceFreq    float64 = 2200.0 // mark frequency in Hertz
	clockPadding         = 5      // extra clock padding at the beginning of the signal for clock recovery
	flagPadding          = 3      // number of flag bytes to send around the message
	mark         symbol  = false
	space        symbol  = true
	zero         byte    = 0
	one          byte    = 1
)

var symbolFrequency = map[symbol]float64{mark: markFreq, space: spaceFreq}

func writeSamples(ax25data []byte, samplesPerSecond uint32, bitsPerSample uint8) {

	ax25data = appendFCS(ax25data)

	currentSymbol := mark

	// send N clock bytes 0x00
	for i := 0; i < clockPadding; i++ {
		for j := 0; j < 8; j++ {
			currentSymbol = nrzi(currentSymbol, 0)
			// write out the samples
		}
	}

	// send M flagBytes
	for i := 0; i < flagPadding; i++ {
		txByte := flagByte
		var bit byte
		for j := 0; j < 8; j++ {
			bit = txByte & 0x01 // get the lowest bit
			currentSymbol = nrzi(currentSymbol, bit)
			// write out the samples
			txByte = txByte >> 1 // shift to the next bit
		}
	}

	// send ax25data with bit-stuffing and update CRC
	consecutiveOnes := 0 // the flag always ends with a zero
	for _, dataByte := range ax25data {
		var bit byte
		for j := 0; j < 8; j++ {
			bit = dataByte & 0x01
			currentSymbol = nrzi(currentSymbol, bit)
			// write out the samples
			if bit == 0 { // update the consecutive ones count
				consecutiveOnes = 0
			} else {
				consecutiveOnes++
			}
			if consecutiveOnes == 5 { // transmit a zero
				currentSymbol = nrzi(currentSymbol, 0)
				// write out the samples
				consecutiveOnes = 0
			}
			dataByte = dataByte >> 1
		}
	}

	// send M flag bytes
	for i := 0; i < flagPadding; i++ {
		txByte := flagByte
		var bit byte
		for j := 0; j < 8; j++ {
			bit = txByte & 0x01
			currentSymbol = nrzi(currentSymbol, bit)
			// write out the samples
			txByte = txByte >> 1
		}
	}
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

func nrzi(currentSymbol symbol, bit byte) symbol {
	// non-return to zero inverted encoding of bit
	// i.e. if bit is 0 switch symbols
	if bit == 1 {
		return currentSymbol
	}
	if currentSymbol == mark {
		return space
	}
	return mark
}
