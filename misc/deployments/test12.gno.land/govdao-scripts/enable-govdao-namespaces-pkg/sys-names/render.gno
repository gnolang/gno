package names

import "gno.land/p/nt/ufmt/v0"

func Render(_ string) string {
	out := "# r/sys/names/v2\n\n"
	out += "Namespace deployment permission registry.\n\n"
	out += "PA namespaces (own address) are always allowed.\n\n"
	if namespaceStore.Size() == 0 {
		out += "_No custom namespaces registered._\n"
		return out
	}
	out += "## Registered Namespaces\n\n"
	out += "| Namespace | Owner |\n|---|---|\n"
	namespaceStore.Iterate("", "", func(key string, val any) bool {
		out += ufmt.Sprintf("| %s | %s |\n", key, val.(address).String())
		return false
	})
	return out
}
