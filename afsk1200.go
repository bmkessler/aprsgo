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

var symbolFrequency = map[symbol]float64{mark: markFreq, space: spaceFreq}

type waveWriter struct {
	currentPhase     float64
	samplesPerSecond uint32
	bitsPerSample    uint8
	data             []byte
}

func (w *waveWriter) WriteSymbol(s symbol) error {
	phaseIncrement := 2 * math.Pi * symbolFrequency[s] / float64(w.samplesPerSecond)
	samplesPerSymbol := w.samplesPerSecond / symbolRate // note the true symbol rate will be off symbolRate * (w.samplesPerSecond % symbolRate)
	for i := uint32(0); i < samplesPerSymbol; i++ {
		w.currentPhase += phaseIncrement
		if w.currentPhase > 2*math.Pi {
			w.currentPhase -= 2 * math.Pi
		}
		newSample := math.Cos(w.currentPhase)
		w.data = append(w.data, byte(newSample))
	}
	return nil
}
