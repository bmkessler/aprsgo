package aprsgo

import (
	"math"
)

type symbolWriter interface {
	WriteSymbol(symbol) error
}

const (
	symbolRate uint32  = 1200   // 1200 baud
	markFreq   float64 = 1200.0 // mark frequency in Hertz
	spaceFreq  float64 = 2200.0 // mark frequency in Hertz
)

var scalingFactor = 0.75 // the scaling factor for volume between 0.0 and 1.0

var symbolFrequency = map[symbol]float64{mark: markFreq, space: spaceFreq}

type waveWriter struct {
	samplesPerSecond     uint32             // ideally a multiple of 1200, e.g. 48000 DVD sound
	bitsPerSample        uint8              // supported values are 8, 12, 32
	volumeLevel          float64            // the scaling factor for the samples to full scale
	symbolCount          uint32             // the count of the current number of symbols in this second
	samplesPerSymbol     uint32             // how many samples are written for each symbol
	skewSamples          uint32             // the skew samples needed to prevent symbol rate drift
	currentPhase         float64            // the current phase of the wave to maintain continuity
	phaseIncrementSymbol map[symbol]float64 // map of the phase increment per sample for each symbol
	data                 []byte             // the output data as an array of bytes
}

func newWaveWriter(samplesPerSecond uint32, bitsPerSample uint8) *waveWriter {
	writer := new(waveWriter)
	writer.samplesPerSecond = samplesPerSecond
	writer.bitsPerSample = bitsPerSample
	writer.volumeLevel = scalingFactor * float64(uint64(1)<<(bitsPerSample-1)) // the level
	writer.samplesPerSymbol = samplesPerSecond / symbolRate
	writer.skewSamples = samplesPerSecond % symbolRate // the remainder will need to be skewed in to prevent drift
	phaseIncrementSymbol := make(map[symbol]float64)
	phaseIncrementSymbol[mark] = 2 * math.Pi * markFreq / float64(samplesPerSecond)
	phaseIncrementSymbol[space] = 2 * math.Pi * spaceFreq / float64(samplesPerSecond)
	writer.phaseIncrementSymbol = phaseIncrementSymbol
	return writer
}

func (w *waveWriter) WriteSymbol(sym symbol) error {
	phaseIncrement := w.phaseIncrementSymbol[sym]
	for i := uint32(0); i < w.samplesPerSymbol; i++ {
		w.writeSample(phaseIncrement)
	}
	if w.symbolCount < w.skewSamples { // write an extra sample in for the first skew samples
		w.writeSample(phaseIncrement)
	}
	w.symbolCount++
	w.symbolCount %= symbolRate // reset the symbol count after every second
	return nil
}

func (w *waveWriter) writeSample(phaseIncrement float64) {
	w.currentPhase += phaseIncrement
	if w.currentPhase > 2*math.Pi { // avoid overflowing the phase
		w.currentPhase -= 2 * math.Pi
	}
	newSample := w.volumeLevel * math.Sin(w.currentPhase)
	switch { // handle different bit rates properly
	case w.bitsPerSample == 8: // 8-bit just write the bytes
		w.data = append(w.data, byte(newSample))
	case w.bitsPerSample == 16: // 16-bit write LSB then MSB
		u16Sample := uint16(newSample)
		var h, l byte = byte((u16Sample >> 8) & 0xff), byte(u16Sample & 0xff)
		w.data = append(w.data, l)
		w.data = append(w.data, h)
	case w.bitsPerSample == 32: // 32-bit write from LSB to MSB
		u32Sample := uint32(newSample)
		var hh, hl, lh, ll byte = byte((u32Sample >> 24) & 0xff), byte((u32Sample >> 16) & 0xff), byte((u32Sample >> 8) & 0xff), byte(u32Sample & 0xff)
		w.data = append(w.data, ll)
		w.data = append(w.data, lh)
		w.data = append(w.data, hl)
		w.data = append(w.data, hh)
	}
}
