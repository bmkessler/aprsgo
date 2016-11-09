package aprsgo

// encoder.go handles encoding of the data, the with Non-Return to Zero Inverted (NRZI)
// bit stuffing and appropriate start and end flags

// Symbol is a physical transmission symbol mark/space
type Symbol bool

// SymbolStream is a series of Symobls for Tx/Rx
type SymbolStream []Symbol

const (
	mark  Symbol = false
	space Symbol = true
	zero  byte   = 0
	one   byte   = 1
)

var (
	clockPadding = 5 // extra clock padding at the beginning of the signal for clock recovery
	flagPadding  = 3 // number of flag bytes to send around the message
)

// Encode converts ax25data from an array of bytes to an array of symbols for transmission
func (ax25data AX25Data) Encode() SymbolStream {
	var symbolStream SymbolStream

	currentSymbol, consecutiveOnes := mark, 0
	// send N clock bytes 0x00
	for i := 0; i < clockPadding; i++ {
		symbolStream, currentSymbol, consecutiveOnes = writeByte(0x00, currentSymbol, consecutiveOnes, false, symbolStream)
	}

	// send M flagBytes
	for i := 0; i < flagPadding; i++ {
		symbolStream, currentSymbol, consecutiveOnes = writeByte(flag, currentSymbol, consecutiveOnes, false, symbolStream)
	}

	// send ax25data with bit-stuffing
	for _, dataByte := range ax25data {
		symbolStream, currentSymbol, consecutiveOnes = writeByte(dataByte, currentSymbol, consecutiveOnes, true, symbolStream)
	}

	// send M flag bytes
	for i := 0; i < flagPadding; i++ {
		symbolStream, currentSymbol, consecutiveOnes = writeByte(flag, currentSymbol, consecutiveOnes, false, symbolStream)
	}
	return symbolStream
}

func nrzi(currentSymbol Symbol, bit byte) Symbol {
	// non-return to zero inverted encoding of bit
	// i.e. if bit is 0 switch symbols
	if bit == 1 {
		return currentSymbol
	}
	return !currentSymbol
}

func writeByte(dataByte byte, currentSymbol Symbol, consecutiveOnes int, bitStuff bool, symbolStream SymbolStream) (SymbolStream, Symbol, int) {
	// iterates over the bits in a byte least-significant first and writes them out
	// returns the symbolStream with any symbols appended
	// the current symbol and count of consecutive ones
	// implements bitstuffing every five ones if bitStuff is true
	var bit byte
	for j := 0; j < 8; j++ {
		bit = dataByte & 0x01
		currentSymbol = nrzi(currentSymbol, bit)
		symbolStream = append(symbolStream, currentSymbol)
		if bit == 0 { // update the consecutive ones count
			consecutiveOnes = 0
		} else {
			consecutiveOnes++
		}
		if bitStuff && consecutiveOnes == 5 { // transmit a zero on 5 consecutive ones if bit-stuffing
			currentSymbol = nrzi(currentSymbol, 0)
			symbolStream = append(symbolStream, currentSymbol)
			consecutiveOnes = 0
		}
		dataByte = dataByte >> 1
	}
	return symbolStream, currentSymbol, consecutiveOnes
}
