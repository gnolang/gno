// wav provides functionality to write WAV (Waveform Audio File Format) files in Gno.
// WAV is a common audio file format used to store audio data with high-quality sound.
package wav

import (
	"io"
	"math"

	"gno.land/p/demo/audio/riff"
)

const (
	// AudioFormatPCM represents the PCM (Pulse Code Modulation) audio format.
	AudioFormatPCM = 1
	// AudioFormatIEEEFloat represents the IEEE floating-point audio format.
	AudioFormatIEEEFloat = 3
	// AudioFormatALaw represents the A-law audio format.
	AudioFormatALaw = 6
	// AudioFormatMULaw represents the μ-law audio format.
	AudioFormatMULaw = 7
)

// WavFormat represents the format information for the WAV audio file.
type WavFormat struct {
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
}

// WavData represents audio data for the WAV file.
type WavData struct {
	io.Reader
	Size uint32
	pos  uint32
}

// Sample represents audio samples.
type Sample struct {
	Values [2]int
}

// Writer allows writing WAV audio data to an underlying io.Writer.
type Writer struct {
	io.Writer
	Format *WavFormat
}

// NewWriter creates a new Writer instance for writing WAV audio data to the provided io.Writer.
// It sets up the necessary WAV format and creates the WAV file header.
func NewWriter(w io.Writer, numSamples uint32, numChannels uint16, sampleRate uint32, bitsPerSample uint16) (writer *Writer, err error) {
	blockAlign := numChannels * bitsPerSample / 8
	byteRate := sampleRate * uint32(blockAlign)
	format := &WavFormat{AudioFormatPCM, numChannels, sampleRate, byteRate, blockAlign, bitsPerSample}
	dataSize := numSamples * uint32(format.BlockAlign)
	riffSize := 4 + 8 + 16 + 8 + dataSize
	riffWriter, err := riff.NewWriter(w, []byte("WAVE"), riffSize)

	writer = &Writer{riffWriter, format}
	riffWriter.WriteChunk([]byte("fmt "), 16)
	riffWriter.WriteUint16(format.AudioFormat)
	riffWriter.WriteUint16(format.NumChannels)
	riffWriter.WriteUint32(format.SampleRate)
	riffWriter.WriteUint32(format.ByteRate)
	riffWriter.WriteUint16(format.BlockAlign)
	riffWriter.WriteUint16(format.BitsPerSample)
	riffWriter.WriteChunk([]byte("data"), dataSize)
	return
}

// WriteSamples writes audio samples to the WAV file.
// It takes a slice of Sample and converts the audio data to the appropriate format before writing it to the file.
func (w *Writer) WriteSamples(samples []Sample) (err error) {
	bitsPerSample := w.Format.BitsPerSample
	numChannels := w.Format.NumChannels

	var i, b uint16
	var by []byte
	for _, sample := range samples {
		for i = 0; i < numChannels; i++ {
			value := toUint(sample.Values[i], int(bitsPerSample))

			for b = 0; b < bitsPerSample; b += 8 {
				by = append(by, uint8((value>>b)&math.MaxUint8))
			}
		}
	}

	_, err = w.Write(by)
	return
}

// toUint converts signed integers to unsigned integers of the specified number of bits.
// This is used when writing samples to the WAV file, as WAV files store audio data in an unsigned format.
func toUint(value int, bits int) uint {
	var result uint

	switch bits {
	case 32:
		result = uint(uint32(value))
	case 16:
		result = uint(uint16(value))
	case 8:
		result = uint(value)
	default:
		if value < 0 {
			result = uint((1 << uint(bits)) + value)
		} else {
			result = uint(value)
		}
	}

	return result
}
