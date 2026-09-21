package gnoweb

import (
	"bytes"
	"testing"
)

// BenchmarkEmphasisStress profiles a single emphasis-stress render against the
// production goldmark instance. Use with:
//
//	go test ./gno.land/pkg/gnoweb/ -run=^$ -bench=BenchmarkEmphasisStress \
//	  -benchtime=1x -cpuprofile=cpu.prof -memprofile=mem.prof -timeout 600s
func BenchmarkEmphasisStress(b *testing.B) {
	sizes := []struct {
		name string
		n    int
	}{
		{"64KB", 1 << 16},
		{"256KB", 1 << 18},
		// The render-input cap: the largest payload that can reach the parser.
		{"1MB-cap", maxMarkdownRenderBytes},
	}
	for _, s := range sizes {
		b.Run(s.name, func(b *testing.B) {
			gm := makeProductionRenderer()
			src := payloadEmphasis(s.n)
			for b.Loop() {
				var buf bytes.Buffer
				if err := gm.Convert(src, &buf); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
