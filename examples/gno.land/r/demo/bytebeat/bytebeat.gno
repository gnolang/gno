package bytebeat

import (
	"gno.land/p/demo/audio/bytebeat"
)

var data string

func init() {
	data = bytebeat.ByteBeat(1, 8000, func(t int) int {
		return (t>>10^t>>11)%5*((t>>14&3^t>>15&1)+1)*t%99 + ((3 + (t >> 14 & 3) - (t >> 16 & 1)) / 3 * t % 99 & 64)
	})
}

func Render(path string) string {
	output := "# bytebeat\n\n"
	output += "(t>>10^t>>11)%5*((t>>14&3^t>>15&1)+1)*t%99 + ((3 + (t >> 14 & 3) - (t >> 16 & 1)) / 3 * t % 99 & 64)\n\n\n"
	output += `<audio controls="controls" autobuffer="autobuffer" autoplay="autoplay">
<source src="data:audio/wav;base64,`
	output += data
	output += `" />
</audio>`
	return output
}
