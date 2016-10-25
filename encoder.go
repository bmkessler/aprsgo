package aprsgo

// This file contains the routines for writing out a WAV file
// of the AFSK1200 data with Non-Return to Zero Inverted (NRZI)
// encoding of the data, bit stuffing and appropriate start and end flags

type symbol bool

const (
	mark  symbol = false
	space symbol = true
	zero  byte   = 0
	one   byte   = 1
)

var (
	clockPadding = 5 // extra clock padding at the beginning of the signal for clock recovery
	flagPadding  = 3 // number of flag bytes to send around the message
)

func writeSamples(ax25data []byte, writer symbolWriter) error {
	var err error
	currentSymbol, consecutiveOnes := mark, 0

	// send N clock bytes 0x00
	for i := 0; i < clockPadding; i++ {
		currentSymbol, consecutiveOnes, err = writeByte(0x00, currentSymbol, consecutiveOnes, false, writer)
		if err != nil {
			return err
		}
	}

	// send M flagBytes
	for i := 0; i < flagPadding; i++ {
		currentSymbol, consecutiveOnes, err = writeByte(flag, currentSymbol, consecutiveOnes, false, writer)
		if err != nil {
			return err
		}
	}

	// send ax25data with bit-stuffing
	for _, dataByte := range ax25data {
		currentSymbol, consecutiveOnes, err = writeByte(dataByte, currentSymbol, consecutiveOnes, true, writer)
		if err != nil {
			return err
		}
	}

	// send M flag bytes
	for i := 0; i < flagPadding; i++ {
		currentSymbol, consecutiveOnes, err = writeByte(flag, currentSymbol, consecutiveOnes, false, writer)
		if err != nil {
			return err
		}
	}
	return nil
}

func nrzi(currentSymbol symbol, bit byte) symbol {
	// non-return to zero inverted encoding of bit
	// i.e. if bit is 0 switch symbols
	if bit == 1 {
		return currentSymbol
	}
	return !currentSymbol
}

func writeByte(dataByte byte, currentSymbol symbol, consecutiveOnes int, bitStuff bool, writer symbolWriter) (symbol, int, error) {
	// iterates over the bits in a byte least-significant first and writes them out
	// returns the current symbol and count of consecutive ones and an error if encountered
	// implements bitstuffing every five ones if bitStuff is true
	var bit byte
	for j := 0; j < 8; j++ {
		bit = dataByte & 0x01
		currentSymbol = nrzi(currentSymbol, bit)
		if err := writer.WriteSymbol(currentSymbol); err != nil {
			return currentSymbol, consecutiveOnes, err
		}
		if bit == 0 { // update the consecutive ones count
			consecutiveOnes = 0
		} else {
			consecutiveOnes++
		}
		if bitStuff && consecutiveOnes == 5 { // transmit a zero on 5 consecutive ones if bit-stuffing
			currentSymbol = nrzi(currentSymbol, 0)
			if err := writer.WriteSymbol(currentSymbol); err != nil {
				return currentSymbol, consecutiveOnes, err
			}
			consecutiveOnes = 0
		}
		dataByte = dataByte >> 1
	}
	return currentSymbol, consecutiveOnes, nil
}
