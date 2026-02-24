package names

func Render(_ string) string {
	return `# r/sys/names
System Realm for checking namespace deployment permissions.

Only address-prefix (PA) namespaces are currently allowed.
Reserved namespaces (gnoland, nt, sys, gov) are blocked.`
}
