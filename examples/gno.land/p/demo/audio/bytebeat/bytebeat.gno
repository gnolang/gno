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

func ByteBeat(seconds uint32, sampleRate uint32, bytebeat_func func(t int) int) (data string) {
	b := new(bytes.Buffer)
	w := bufio.NewWriter(b)

	var numSamples uint32 = sampleRate * seconds / 2
	var numChannels uint16 = 1
	var bitsPerSample uint16 = 16

	writer, err := wav.NewWriter(w, numSamples*2, numChannels, sampleRate, bitsPerSample)
	if err != nil {
		data = ufmt.Sprintf("error: %s", err)
		return
	}

	samples := make([]wav.Sample, numSamples)
	maxValue := 0
	lastValue := 0
	for i := int(0); i < int(numSamples); i++ {
		lastValue = bytebeat_func(i) % 255
		samples[i].Values[0] = lastValue
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
