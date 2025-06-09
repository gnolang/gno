package echo

/*
 * This realm echoes the `path` argument it received.
 * Can be used by developers as a simple endpoint to test
 * forbidden characters, for pentesting or simply to
 * test it works.
 *
 * See also r/demo/print (to print various thing like user address)
 */
func Render(path string) string {
	return path
}
