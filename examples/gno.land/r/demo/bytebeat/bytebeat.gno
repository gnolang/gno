package bytebeat

import (
	"gno.land/p/demo/audio/bytebeat"
	evalint32 "gno.land/p/demo/math_eval/int32"
	"gno.land/p/demo/ufmt"
)

func Render(path string) string {
	output := "# bytebeat\n\n"
	if path == "" {
		u := "/r/demo/bytebeat:(t>>6|t|t>>(t>>16))*10+((t>>11)&7)"
		output += "generate audio by entering a expression with the variable *t*:\n\n"
		output += ufmt.Sprintf("[%s](%s)", u, u)
	} else {
		exp, err := evalint32.Parse(path)
		if err != nil {
			return err.Error()
		}
		variables := make(map[string]int)
		data := bytebeat.ByteBeat(1, 8000, func(t int) int {
			variables["t"] = t
			res, _ := evalint32.Eval(exp, variables)
			return res
		})
		output += "```\n" + path + "\n```\n\n\n"
		output += `<audio controls="controls" autobuffer="autobuffer" autoplay="autoplay">
	<source src="data:audio/wav;base64,`
		output += data
		output += `" />
	</audio>`
	}
	return output
}
