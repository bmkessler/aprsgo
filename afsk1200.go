package aprsgo

import (
	"math"
)

type symbolWriter interface {
	WriteSymbol(symbol) error
}

const (
	symbolRate    uint32  = 1200   // 1200 baud
	markFreq      float64 = 1200.0 // mark frequency in Hertz
	spaceFreq     float64 = 2200.0 // mark frequency in Hertz
	scalingFactor float64 = 0.75   // the scaling factor for volume
)

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
	writer.volumeLevel = scalingFactor * float64(uint64(1)<<(bitsPerSample-1)) //
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
		w.currentPhase += phaseIncrement
		if w.currentPhase > 2*math.Pi {
			w.currentPhase -= 2 * math.Pi
		}
		newSample := w.volumeLevel * math.Cos(w.currentPhase)
		w.data = append(w.data, byte(newSample)) // handle different bit rates properly
	}
	if w.symbolCount < w.skewSamples { // write an extra sample in for the firsr skew samples
		w.currentPhase += phaseIncrement
		if w.currentPhase > 2*math.Pi {
			w.currentPhase -= 2 * math.Pi
		}
		newSample := w.volumeLevel * math.Cos(w.currentPhase)
		w.data = append(w.data, byte(newSample)) // handle different bit rates properly
	}
	w.symbolCount++
	w.symbolCount %= symbolRate
	return nil
}
