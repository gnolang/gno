package wav

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"testing"
)

func TestWav(t *testing.T) {
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)

	var numSamples uint32 = 2
	var numChannels uint16 = 2
	var sampleRate uint32 = 44100
	var bitsPerSample uint16 = 16

	writer, err := NewWriter(foo, numSamples, numChannels, sampleRate, bitsPerSample)
	if err != nil {
		t.Fatal(err)
	}
	samples := make([]Sample, numSamples)

	samples[0].Values[0] = 32767
	samples[0].Values[1] = -32768
	samples[1].Values[0] = 123
	samples[1].Values[1] = -123

	err = writer.WriteSamples(samples)
	if err != nil {
		t.Fatal(err)
	}

	foo.Flush()
	output := base64.StdEncoding.EncodeToString(b.Bytes())
	if output != "UklGRiwAAABXQVZFZm10IBAAAAABAAIARKwAABCxAgAEABAAZGF0YQgAAAD/fwCAewCF/w==" {
		t.Errorf("wrong output: %s", output)
	}
}
