package aprsgo

import (
	"testing"
)

func TestNewWaveWriter(t *testing.T) {
	wr := newWaveWriter(48000, 8)
	if wr.samplesPerSymbol != 40 {
		t.Errorf("48kHz sampling should have 40 samples per symbol, has %v", wr.samplesPerSymbol)
	}
	if wr.skewSamples != 0 {
		t.Errorf("48kHz sampling should have no skew samples, has %v", wr.skewSamples)
	}
	if wr.volumeLevel != 96 {
		t.Errorf("8-bit volume level with scaling 0.75 should be XXX, got %v", int8(wr.volumeLevel))
	}
	wr = newWaveWriter(44100, 16)
	if wr.samplesPerSymbol != 36 {
		t.Errorf("44.1kHz sampling should have 36 samples per symbol, has %v", wr.samplesPerSymbol)
	}
	if wr.skewSamples != 900 {
		t.Errorf("44.1kHz sampling should have 900 skew samples, has %v", wr.skewSamples)
	}
	if wr.volumeLevel != 24576 {
		t.Errorf("16-bit volume level with scaling 0.75 should be 24576, got %v", int16(wr.volumeLevel))
	}
}
