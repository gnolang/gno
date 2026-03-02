package cla

import "gno.land/p/moul/helplink"

func Render(path string) string {
	if requiredHash == "" {
		return "# Gno CLA Registry\n\n**Status:** CLA enforcement is DISABLED\n"
	}

	output := "# Gno CLA Registry\n\n**Status:** CLA enforcement is ENABLED\n\n"
	output += "**Required Hash:** " + requiredHash + "\n"
	if claURL != "" {
		output += "**CLA Document:** " + claURL + "\n"
	}
	output += "\n### Actions\n\n"
	output += helplink.Func("Sign CLA", "Sign", "hash", requiredHash) + "\n"
	return output
}
