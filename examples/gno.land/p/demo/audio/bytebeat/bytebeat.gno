// bytebeat provides utilities for generating audio data using bytebeat
// functions and encoding the data as WAV audio files.
package bytebeat

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"math"

	"gno.land/p/demo/audio/biquad"
	"gno.land/p/demo/audio/wav"
	"gno.land/p/demo/ufmt"
)

// ByteBeat takes in several parameters to generate audio data based on a bytebeat function.
// Bytebeat is a technique for generating audio using simple mathematical expressions that operate on the time variable.
// The generated audio data is returned as a base64-encoded string representing a WAV audio file.

// Inputs:
// - seconds (uint32): The duration of the generated audio in seconds.
// - sampleRate (uint32): The sample rate of the generated audio in samples per second (Hz).
// - bytebeat_func (func(t int) int): The bytebeat function that defines the audio generation.
//   It takes an integer t as input (representing time in samples) and returns an integer value that represents the amplitude of the audio signal at that specific time.

// Outputs:
// - data (string): The generated audio data encoded as a base64 string in WAV format.
//   The audio data represents a mono audio stream with 16 bits per sample.

// Function Logic:
// 1. The function initializes a buffer (b) and a buffered writer (w) for writing the WAV audio data.
// 2. It calculates the total number of samples (numSamples) based on the provided duration and sample rate.
// 3. It sets the number of channels to 1 (mono) and the bits per sample to 16.
// 4. The function creates a new WAV writer with the provided specifications (sample rate, number of channels, and bits per sample) using the buffer writer.
// 5. It generates the audio samples by iterating through numSamples using the bytebeat function (bytebeat_func).
//    The bytebeat function returns an integer value, which is then used as the amplitude value for the current sample.
// 6. The function normalizes the generated samples by scaling them to the range [-28000, 28000].
// 7. It applies a highpass filter with a cutoff frequency of 10 Hz to remove any DC bias in the audio data.
// 8. The function writes the generated samples to the WAV writer.
// 9. After writing the audio data, it flushes the buffered writer to ensure all data is written to the underlying buffer.
// 10. Finally, the function returns the base64-encoded string representation of the generated audio data.

// Note: If any errors occur during the process (e.g., WAV writer initialization, writing samples), the function will set data to an error message indicating the issue encountered. Otherwise, data will contain the base64-encoded WAV audio data.
func ByteBeat(seconds uint32, sampleRate uint32, bytebeat_func func(t int) int) (data string) {
	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)

	numSamples := sampleRate * seconds / 2
	numChannels := uint16(1)
	bitsPerSample := uint16(16)

	writer, err := wav.NewWriter(w, numSamples*2, numChannels, sampleRate, bitsPerSample)
	if err != nil {
		data = ufmt.Sprintf("error: %s", err)
		return
	}

	samples := make([]wav.Sample, numSamples)
	maxValue := 0
	for i := 0; i < int(numSamples); i++ {
		samples[i].Values[0] = bytebeat_func(i) % 255
		if samples[i].Values[0] > maxValue {
			maxValue = samples[i].Values[0]
		}
	}

	// normalize samples
	for i := uint32(0); i < numSamples; i++ {
		samples[i].Values[0] = (samples[i].Values[0] * 28000 / maxValue)
	}
	// highpass filter around 10 hz.
	// this is used to remove the DC bias inherit in the gneeration of the
	// bytebeats because speakers like values in the range [-1,1]
	// rather than in the range [0,2]
	// More information is available here: https://en.wikipedia.org/wiki/DC_bias
	hpf := biquad.New(10, float64(sampleRate), 0.707, 0, "highpass")
	for i := uint32(0); i < numSamples; i++ {
		samples[i].Values[0] = int(math.Round(hpf.Update(float64(samples[i].Values[0]))))
	}

	err = writer.WriteSamples(samples)
	if err != nil {
		data = ufmt.Sprintf("error: %s", err)
		return
	}

	err = writer.WriteSamples(samples)
	if err != nil {
		data = ufmt.Sprintf("error: %s", err)
		return
	}

	w.Flush()

	data = base64.StdEncoding.EncodeToString(b.Bytes())
	return
}
