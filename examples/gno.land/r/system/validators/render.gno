package validators

import (
	"gno.land/p/demo/ufmt"
)

func Render(path string) string {
	output := ""
	output += ufmt.Sprintf("## Validators: (%d / %d)\n", validators.Size(), maxValidators)
	output += "\n"
	output += "| Name | Address | VotingPower |\n"
	output += "| ---- | ------- | ----------- |\n"

	validators.Iterate("", "", func(key string, value interface{}) bool {
		v := value.(*Validator)
		output += v.Render()
		output += "\n"
		return false
	})

	return output
}
