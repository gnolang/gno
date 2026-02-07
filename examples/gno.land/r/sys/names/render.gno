package names

func Render(_ string) string {
	return `# r/sys/names
System Realm for checking namespace deployment permissions.

Namespace enforcement is controlled by the VM param ` + "`sysnames_enabled`" + `.
When enabled, only the registered owner of a namespace (or users deploying to their own address) may publish packages under that namespace.`
}
