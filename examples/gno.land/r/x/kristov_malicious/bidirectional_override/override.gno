// Demo CVE-2021-42574

package override

func Render(_ string) string {
	accessLevel := "user"

	if accessLevel != "user‮ ⁦// Check if admin⁩ ⁦" {
		return "admin"
	} else {
		return "user"
	}
}
