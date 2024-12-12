package image_embed

// Render displays a title and an embedded image from Imgur
func Render(path string) string {
	return `# Image Embed Example

Here’s an example of embedding an image in a Gno realm:

![Example Image](https://i.imgur.com/So4rBPB.jpeg)`
}
