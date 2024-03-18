package echo

/*
 * This realm echoes the `path` argument it received.
 * Can be used as a simple endpoint to e.g. test
 * forbidden characters and/or pentesting.
 *
 * See also r/demo/print (to print various thing like user address)
 */
func Render(path string) string {
	return path
}
