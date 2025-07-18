// MDPack demonstrates embedding files in markdown using txtar-like format
package mdpack

func Render(string) string {
	return `# MDPack Example

Embed files in markdown with optional metadata.

Here's an image: ![Gno Logo](logo.png) ([download](?file=logo.png))

And a data file: ([raw](?file=data.json))

` + "```pack,image.jpg?encoding=base64&mimetype=image/jpeg --" + `
iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=
` + "```" + `

` + "```pack,data.json?embed=true --" + `
{"foo": "bar"}
` + "```"
}
