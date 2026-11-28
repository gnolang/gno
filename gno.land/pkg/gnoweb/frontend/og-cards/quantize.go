//go:build ignore

// quantize.go rewrites a grayscale PNG with 16 gray levels, a 4-bit palette
// that keeps text edges smooth at a third of the size.
//
// Usage: go run quantize.go <in.png> <out.png>
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func main() {
	in, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	src, err := png.Decode(in)
	if err != nil {
		log.Fatal(err)
	}
	palette := make(color.Palette, 16)
	for i := range palette {
		palette[i] = color.Gray{Y: uint8(i * 17)}
	}
	dst := image.NewPaletted(src.Bounds(), palette)
	b := src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	if err := (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(out, dst); err != nil {
		log.Fatal(err)
	}
	if err := out.Close(); err != nil {
		log.Fatal(err)
	}
}
