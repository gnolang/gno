package tnft

import (
	b64 "encoding/base64"
	"math/rand"
	"strings"

	"gno.land/p/nt/ufmt"
)

var baseTempalte = `<svg width="135" height="135" viewBox="0 0 135 135" fill="none" xmlns="http://www.w3.org/2000/svg">
<g clip-path="url(#clip0_7698_56846)">
<circle cx="67.5" cy="67.5" r="67.5" fill="url(#paint0_linear_7698_56846)"/>
<path d="M51.2905 42.9449L66.4895 33L97 52.8061L81.8241 62.7425L51.2905 42.9449Z" fill="white"/>
<path d="M51.6055 67.5059L66.8044 57.561L97 77.0657L82.1046 87.1793L51.6055 67.5059Z" fill="white" fill-opacity="0.4"/>
<path d="M36.0464 81.7559L51.2905 71.811L81.7336 91.6547L66.4895 101.508L36.0464 81.7559Z" fill="white" fill-opacity="0.6"/>
<path d="M36.001 52.8055L51.2884 42.9177L51.2884 71.8145L36.001 81.779L36.001 52.8055Z" fill="white"/>
<path d="M82.1051 87.1797L97.0016 77.0662L97.0016 81.7029L81.7896 91.629L82.1051 87.1797Z" fill="white" fill-opacity="0.5"/>
<path d="M51.2905 42.9449L66.4895 33L97 52.8061L81.8241 62.7425L51.2905 42.9449Z" fill="white"/>
<path d="M51.6055 67.5059L66.8044 57.561L97 77.0657L82.1046 87.1793L51.6055 67.5059Z" fill="white" fill-opacity="0.4"/>
<path d="M36.0464 81.7559L51.2905 71.811L81.7336 91.6547L66.4895 101.508L36.0464 81.7559Z" fill="white" fill-opacity="0.6"/>
<path d="M36.001 52.8055L51.2884 42.9177L51.2884 71.8145L36.001 81.779L36.001 52.8055Z" fill="white"/>
<path d="M82.1051 87.1797L97.0016 77.0662L97.0016 81.7029L81.7896 91.629L82.1051 87.1797Z" fill="white" fill-opacity="0.5"/>
</g>
<defs>
<linearGradient id="paint0_linear_7698_56846" x1="%d" y1="%d" x2="%d" y2="%d" gradientUnits="userSpaceOnUse">
<stop stop-color="%s"/>
<stop offset="1" stop-color="%s"/>
</linearGradient>
<clipPath id="clip0_7698_56846">
<rect width="135" height="135" fill="white"/>
</clipPath>
</defs>
</svg>
`

// charset contains valid hex digits for color generation.
const charset = "0123456789ABCDEF"

// genImageURI generates a base64-encoded SVG image URI with random gradient colors.
func genImageURI(r *rand.Rand) string {
	imageRaw := genImageRaw(r)
	sEnc := b64.StdEncoding.EncodeToString([]byte(imageRaw))

	return "data:image/svg+xml;base64," + sEnc
}

// genImageRaw generates an SVG image with random gradient parameters.
func genImageRaw(r *rand.Rand) string {
	x1 := 7 + r.Uint64N(7)
	y1 := 7 + r.Uint64N(7)

	x2 := 121 + r.Uint64N(6)
	y2 := 121 + r.Uint64N(6)

	var color1, color2 strings.Builder
	color1.Grow(7)
	color2.Grow(7)
	color1.WriteByte('#')
	color2.WriteByte('#')

	for i := 0; i < 6; i++ {
		color1.WriteByte(charset[r.IntN(16)])
		color2.WriteByte(charset[r.IntN(16)])
	}

	randImage := ufmt.Sprintf(baseTempalte, x1, y1, x2, y2, color1.String(), color2.String())
	return randImage
}
