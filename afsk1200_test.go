package aprsgo

import (
	"fmt"
	"testing"
)

func TestNewWaveWriter(t *testing.T) {
	wr, err := NewWaveWriter(48000, 8, 1)
	if err != nil {
		t.Errorf("Errors creating 48kHz sampling at 8-bits per second: %v", err)
	}
	if wr.samplesPerSymbol != 40 {
		t.Errorf("48kHz sampling should have 40 samples per symbol, has %v", wr.samplesPerSymbol)
	}
	if wr.skewSamples != 0 {
		t.Errorf("48kHz sampling should have no skew samples, has %v", wr.skewSamples)
	}
	if wr.volumeLevel != 96 {
		t.Errorf("8-bit volume level with scaling 0.75 should be XXX, got %v", int8(wr.volumeLevel))
	}
	wr, err = NewWaveWriter(44100, 16, 1)
	if wr.samplesPerSymbol != 36 {
		t.Errorf("44.1kHz sampling should have 36 samples per symbol, has %v", wr.samplesPerSymbol)
	}
	if wr.skewSamples != 900 {
		t.Errorf("44.1kHz sampling should have 900 skew samples, has %v", wr.skewSamples)
	}
	if wr.volumeLevel != 24576 {
		t.Errorf("16-bit volume level with scaling 0.75 should be 24576, got %v", int16(wr.volumeLevel))
	}
	wr, err = NewWaveWriter(44100, 12, 1)
	if err == nil {
		t.Errorf("Created an unsupported wave writer with 12-bits per sample")
	}
	wr, err = NewWaveWriter(44100, 64, 1)
	if err == nil {
		t.Errorf("Created an unsupported wave writer with 64-bits per sample")
	}
}

func TestWriteSymbol(t *testing.T) {
	sampleRate, bitsPerSample, channels := uint32(48000), uint8(8), uint8(1)
	wr, err := NewWaveWriter(sampleRate, bitsPerSample, channels)
	if err != nil {
		t.Errorf("Errors creating %vkHz sampling at %v-bits per second: %v", sampleRate/1000, bitsPerSample, err)
	}
	for i := uint32(0); i < symbolRate; i++ { // write one second of marks
		wr.WriteSymbol(mark)
	}
	if len(wr.data) != int(sampleRate)*int(channels*bitsPerSample)/8 {
		t.Errorf("%v samples, should be %v", len(wr.data), int(sampleRate)*int(bitsPerSample)/8)
	}
	err = wr.WriteFile("test_48000Hz_8bit_1200_mark.wav")
	if err != nil {
		t.Errorf("Error saving file: %v", err)
	}

	sampleRate, bitsPerSample = uint32(44100), uint8(16)
	wr, err = NewWaveWriter(sampleRate, bitsPerSample, channels)
	if err != nil {
		t.Errorf("Errors creating %vkHz sampling at %v-bits per second: %v", sampleRate/1000, bitsPerSample, err)
	}
	for i := uint32(0); i < symbolRate; i++ { // write one second of spaces
		wr.WriteSymbol(space)
	}
	if len(wr.data) != int(sampleRate)*int(channels*bitsPerSample)/8 {
		t.Errorf("%v samples, should be %v", len(wr.data), int(sampleRate)*int(bitsPerSample)/8)
	}
	err = wr.WriteFile("test_44100Hz_16bit_1200_space.wav")
	if err != nil {
		t.Errorf("Error saving file: %v", err)
	}

	sampleRate, bitsPerSample, channels = uint32(48000), uint8(16), 2
	wr, err = NewWaveWriter(sampleRate, bitsPerSample, channels)
	if err != nil {
		t.Errorf("Errors creating %vkHz sampling at %v-bits per second: %v", sampleRate/1000, bitsPerSample, err)
	}
	for i := uint32(0); i < symbolRate; i++ { // write one second of marks in stereo
		wr.WriteSymbol(mark)
	}
	if len(wr.data) != int(sampleRate)*int(channels*bitsPerSample)/8 {
		t.Errorf("%v samples, should be %v", len(wr.data), int(sampleRate)*int(bitsPerSample)/8)
	}
	err = wr.WriteFile(fmt.Sprintf("test_%vHz_%vbit_1200_mark_%vchan.wav", sampleRate, bitsPerSample, channels))
	if err != nil {
		t.Errorf("Error saving file: %v", err)
	}
}

func TestWriteFile(t *testing.T) {
	sampleRate, bitsPerSample, channels := uint32(48000), uint8(16), uint8(1)
	wr, err := NewWaveWriter(sampleRate, bitsPerSample, channels)
	if err != nil {
		t.Errorf("Errors creating %vkHz sampling at %v-bits per second: %v", sampleRate/1000, bitsPerSample, err)
	}
	symbol := mark
	for i := uint32(0); i < symbolRate; i++ { // write one second of alternating marks and spaces
		wr.WriteSymbol(symbol)
		symbol = !symbol
	}
	if len(wr.data) != int(sampleRate)*int(channels*bitsPerSample)/8 {
		t.Errorf("%v samples, should be %v", len(wr.data), int(sampleRate)*int(bitsPerSample)/8)
	}
	err = wr.WriteFile(fmt.Sprintf("test_%vHz_%vbit_1200_clock_%vchan.wav", sampleRate, bitsPerSample, channels))
	if err != nil {
		t.Errorf("Error saving file: %v", err)
	}
}
