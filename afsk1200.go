package aprsgo

// afsk1200.go contains the routines for writing out a WAV file
// of the AFSK1200 data at various bit and sample rates

import (
	"encoding/binary"
	"errors"
	"math"
	"os"
)

// WAVParams are parameters for writing a WAV file
type WAVParams struct {
	Filename         string
	SamplesPerSecond uint32
	BitsPerSample    uint8
	NumChannels      uint8
}

type symbolWriter interface {
	WriteSymbol(Symbol) error
}

const (
	symbolRate uint32  = 1200   // 1200 baud
	markFreq   float64 = 1200.0 // mark frequency in Hertz
	spaceFreq  float64 = 2200.0 // mark frequency in Hertz
)

var (
	riffTag = "RIFF" // RIFF tag header for entire file
	waveTag = "WAVE" // WAVE tag header identifying type of RIFF
	fmtTag  = "fmt " // fmt tag header for format chunk
	dataTag = "data" // data tag header for data chunk
)

var scalingFactor = 0.75 // the scaling factor for volume between 0.0 and 1.0

// SetVolume sets the scaling factor for volume between 0.0 and 1.0
func SetVolume(newVolume float64) {
	if newVolume > 1.0 { // clipped to range
		newVolume = 1.0
	}
	if newVolume < 0.0 {
		newVolume = 0.0
	}
	scalingFactor = newVolume
}

// WriteWAV writes a symbol stream out to a WAV file with parameters given by params
func WriteWAV(symbolStream []Symbol, params WAVParams) error {
	wr, err := newWaveWriter(params.SamplesPerSecond, params.BitsPerSample, params.NumChannels)
	if err != nil {
		return err
	}
	for _, sym := range symbolStream { // encode the symbols
		wr.writeSymbol(sym)
	}
	if err = wr.writeFile(params.Filename); err != nil {
		return err
	}

	return nil
}

var symbolFrequency = map[Symbol]float64{mark: markFreq, space: spaceFreq}

type waveWriter struct {
	samplesPerSecond     uint32             // ideally a multiple of 1200, e.g. 48000 DVD sound
	bitsPerSample        uint8              // supported values are 8, 12, 32
	numChannels          uint8              // 1 mono, 2 stereo
	volumeLevel          float64            // the scaling factor for the samples to full scale
	symbolCount          uint32             // the count of the current number of symbols in this second
	samplesPerSymbol     uint32             // how many samples are written for each symbol
	skewSamples          uint32             // the skew samples needed to prevent symbol rate drift
	currentPhase         float64            // the current phase of the wave to maintain continuity
	phaseIncrementSymbol map[Symbol]float64 // map of the phase increment per sample for each symbol
	data                 []byte             // the output data as an array of bytes
}

type waveHeader struct {
	riffChunkID           [4]byte // "RIFF"
	riffChunkSize         uint32  // 4 + (8 + formatChunkSize) + (8 + dataChunkSize) = 36 + dataChunkSize
	waveChunkID           [4]byte // "WAVE"
	formatChunkID         [4]byte // "fmt "
	formatChunkSize       uint32  // 16 for PCM
	waveFormatTag         uint16  // 0x0001 for PCM
	numberOfChannels      uint16  // Nc
	samplesPerSecond      uint32  // sampling frequency, e.g. 48000
	averageBytesPerSecond uint32  // F*M*Nc
	blockAlign            uint16  // M*Nc
	bitsPerSample         uint16  // 8*M
	dataChunkID           [4]byte // "data"
	dataChunkSize         uint32  // M*Nc*Ns
}

func newWaveWriter(samplesPerSecond uint32, bitsPerSample uint8, numChannels uint8) (*waveWriter, error) {
	if bitsPerSample%8 != 0 || bitsPerSample > 32 {
		return nil, errors.New("only 8, 16, 24 and 32 bitsPerSample are supported")
	}
	writer := new(waveWriter)
	writer.samplesPerSecond = samplesPerSecond
	writer.bitsPerSample = bitsPerSample
	writer.numChannels = numChannels
	writer.volumeLevel = scalingFactor * float64(uint64(1)<<(bitsPerSample-1)) // the level
	writer.samplesPerSymbol = samplesPerSecond / symbolRate
	writer.skewSamples = samplesPerSecond % symbolRate // the remainder will need to be skewed in to prevent drift
	phaseIncrementSymbol := make(map[Symbol]float64)
	phaseIncrementSymbol[mark] = 2 * math.Pi * markFreq / float64(samplesPerSecond)
	phaseIncrementSymbol[space] = 2 * math.Pi * spaceFreq / float64(samplesPerSecond)
	writer.phaseIncrementSymbol = phaseIncrementSymbol
	return writer, nil
}

func (w *waveWriter) writeSymbol(sym Symbol) {
	phaseIncrement := w.phaseIncrementSymbol[sym]
	for i := uint32(0); i < w.samplesPerSymbol; i++ {
		w.writeSample(phaseIncrement)
	}
	if w.symbolCount < w.skewSamples { // write an extra sample in for the first skew samples
		w.writeSample(phaseIncrement)
	}
	w.symbolCount++
	w.symbolCount %= symbolRate // reset the symbol count after every second
}

func (w *waveWriter) writeSample(phaseIncrement float64) {
	w.currentPhase += phaseIncrement
	if w.currentPhase > 2*math.Pi { // avoid overflowing the phase
		w.currentPhase -= 2 * math.Pi
	}
	newSample := w.volumeLevel * math.Sin(w.currentPhase)
	for i := uint8(0); i < w.numChannels; i++ { // write one sample for each channel
		u32Sample := uint32(newSample) // bits per sample only supported multiples of 8 up to 32
		if w.bitsPerSample == 8 {
			u32Sample += (1 << 7) // 8-bit is offset encoded
			u32Sample %= (1 << 8)
		}
		for i := uint8(0); i < w.bitsPerSample/8; i++ {
			w.data = append(w.data, byte(u32Sample&0xFF))
			u32Sample = u32Sample >> 8
		}
	}
}

func (w *waveWriter) writeFile(filename string) error {
	data := w.data
	if len(w.data)%2 != 0 {
		data = append(data, byte(0)) // pad a zero byte if the length is not even
	}
	M := w.bitsPerSample / 8      // Bytes per sample
	Nc := w.numChannels           // number of channels
	dataSize := uint32(len(data)) // the total number of bytes, with padding
	header := waveHeader{
		riffChunkSize:         uint32(36 + dataSize),
		formatChunkSize:       uint32(16),
		waveFormatTag:         uint16(0x0001),
		numberOfChannels:      uint16(Nc),
		samplesPerSecond:      uint32(w.samplesPerSecond),
		averageBytesPerSecond: w.samplesPerSecond * uint32(M) * uint32(Nc),
		blockAlign:            uint16(M) * uint16(Nc),
		bitsPerSample:         uint16(w.bitsPerSample),
		dataChunkSize:         dataSize,
	}
	copy(header.riffChunkID[:], riffTag)
	copy(header.waveChunkID[:], waveTag)
	copy(header.formatChunkID[:], fmtTag)
	copy(header.dataChunkID[:], dataTag)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	err = binary.Write(file, binary.LittleEndian, header)
	if err != nil {
		return err
	}
	err = binary.Write(file, binary.LittleEndian, w.data)
	if err != nil {
		return err
	}

	return nil
}
