package validators

import (
	"gno.land/p/demo/ufmt"
)

func Render(path string) string {
	output := ""
	output += ufmt.Sprintf("## Validators: (%d)\n", validators.Size())
	output += "\n"
	output += "| Name | Address | Pubkey | VotingPower |\n"
	output += "| ---- | ------- | ------ | ----------- |\n"

	validators.Iterate("", "", func(key string, value interface{}) bool {
		v := value.(*Validator)
		output += v.Render()
		output += "\n"
		return false
	})

	return output
}
