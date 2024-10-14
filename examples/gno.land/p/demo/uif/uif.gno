package uif

import (
	"strconv"
	"strings"
)

// Theme structure and default theme
type Theme struct {
	PrimaryColor   string
	SecondaryColor string
	FontSize       string
	BorderRadius   string
	DefaultPadding string
	DefaultMargin  string
}

func DefaultTheme() Theme {
	return Theme{
		PrimaryColor:   "#3498db",   // Blue color
		SecondaryColor: "#2ecc71",   // Green color
		FontSize:       "18px",      // Slightly larger font size
		BorderRadius:   "8px",       // More rounded corners
		DefaultPadding: "12px 24px", // Custom padding for better spacing
		DefaultMargin:  "10px",      // Default margin
	}
}

// Theme method for Navbar
func (theme Theme) Navbar(links map[string]string) string {
	nav := `<nav style='background-color: ` + theme.PrimaryColor + `;
padding: 10px;
display: flex;
justify-content: space-around;
align-items: center;
'>`
	for text, href := range links {
		nav += `<a href='` + href + `' style='color: white; text-decoration: none; padding: 10px 15px;'>` + text + `</a>`
	}
	nav += `</nav>`
	return nav
}

// Margin wraps content with customizable margin
func (theme Theme) Margin(content string, margin string) string {
	if margin == "" {
		margin = theme.DefaultMargin // Use default margin
	}
	return `<div style='margin: ` + margin + `;'>` + content + `</div>`
}

// Padding wraps content with customizable padding
func (theme Theme) Padding(content string, padding string) string {
	if padding == "" {
		padding = theme.DefaultPadding // Use default padding
	}
	return `<div style='padding: ` + padding + `;'>` + content + `</div>`
}

// Theme method for Button
func (theme Theme) Button(text, href string) string {
	return "\n" + `<a href='` + href + `' style='
background: linear-gradient(to right, ` + theme.PrimaryColor + `, ` + theme.SecondaryColor + `); /* Gradient */
border: none;
color: white;
padding: ` + theme.DefaultPadding + `;
text-align: center;
text-decoration: none;
display: inline-block;
font-size: ` + theme.FontSize + `;
margin-top: ` + theme.DefaultMargin + `;
cursor: pointer;
border-radius: ` + theme.BorderRadius + `;
transition: background 0.3s ease;
'>` + text + `</a>` + "\n"
}

// Theme method for Image with customizable width and height
func (theme Theme) Image(src, width, height string) string {
	// Set default width and height if not provided
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "auto"
	}
	return `<img src='` + src + `' alt='Sample Image' style='
width: ` + width + `;
height: ` + height + `;
border-radius: ` + theme.BorderRadius + `;
margin-bottom: ` + theme.DefaultMargin + `;
' />`
}

// Theme method for Card with customizable width and height
func (theme Theme) Card(content, width, height string) string {
	// Set default width and height if not provided
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "auto"
	}
	return `<div style='
width: ` + width + `;
height: ` + height + `;
border: 1px solid #ddd;
margin: ` + theme.DefaultMargin + `;
border-radius: ` + theme.BorderRadius + `;
padding: ` + theme.DefaultPadding + `;
box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
'>` + content + `</div>`
}

// Theme method for TextInput with name and placeholder
func (theme Theme) TextInput(name, placeholder string) string {
	return `<div style='margin-bottom: ` + theme.DefaultMargin + `;'>
<label for='` + name + `' style='
display: block;
margin-bottom: 5px;
font-size: ` + theme.FontSize + `;
color: ` + theme.PrimaryColor + `;
font-weight: bold;
'>` + strings.Title(name) + `:</label>
<input type='text' id='` + name + `' name='` + name + `' placeholder='` + placeholder + `' style='
width: 100%;
padding: ` + theme.DefaultPadding + `;
border-radius: ` + theme.BorderRadius + `;
border: 1px solid #ccc;
font-size: ` + theme.FontSize + `;
color: #333;
transition: border-color 0.3s ease, box-shadow 0.3s ease;
outline: none;
' onfocus='this.style.borderColor="` + theme.PrimaryColor + `"; this.style.boxShadow="0 0 5px ` + theme.PrimaryColor + `";'
onblur='this.style.borderColor="#ccc"; this.style.boxShadow="0 1px 3px rgba(0, 0, 0, 0.1)";'
>
</div>`
}

// Theme method for CodeBlock
func (theme Theme) CodeBlock(code string) string {
	return `<pre style='
background-color: #f5f5f5;
color: #333;
padding: ` + theme.DefaultPadding + `;
border-radius: ` + theme.BorderRadius + `;
overflow-x: auto;
font-size: ` + theme.FontSize + `;
margin: ` + theme.DefaultMargin + `;
'><code>` + code + `</code></pre>
`
}

// Theme method for Grid
func (theme Theme) Grid(items []string, columns int, gap string, textAlign string) string {
	columnsStr := strconv.Itoa(columns)
	grid := `<div style='
display: grid;
grid-template-columns: repeat(` + columnsStr + `, 1fr);
gap: ` + gap + `;
'>`
	for _, item := range items {
		grid += `<div style='text-align: ` + textAlign + `;'>` + item + `</div>`
	}
	grid += `</div>`
	return grid
}

// Theme method for Footer
func (theme Theme) Footer(links []string) string {
	footer := `<footer style='
background-color: ` + theme.SecondaryColor + `;
padding: 10px;
text-align: center;
margin-top: 50px;
'>`
	for _, link := range links {
		footer += `<a href='#' style='color: white; text-decoration: none; padding: 10px 15px;'>` + link + `</a>`
	}
	footer += `</footer>`
	return footer
}

// Theme method for Heading with adjustable size and margin based on level
func (theme Theme) Heading(level int, text string) string {
	if level < 1 || level > 6 {
		level = 1
	}
	tag := "h" + strconv.Itoa(level)
	var fontSize, margin string
	switch level {
	case 1:
		fontSize = "34px"
		margin = "24px 0"
	case 2:
		fontSize = "30px"
		margin = "18px 0"
	case 3:
		fontSize = "24px"
		margin = "16px 0"
	case 4:
		fontSize = "20px"
		margin = "14px 0"
	case 5:
		fontSize = "18px"
		margin = "12px 0"
	case 6:
		fontSize = "16px"
		margin = "10px 0" // Smallest font for h6
	default:
		fontSize = theme.FontSize
		margin = theme.DefaultMargin
	}
	return `<` + tag + ` style='
font-size: ` + fontSize + `;
color: ` + theme.PrimaryColor + `;
margin: ` + margin + `;
'>` + text + `</` + tag + `>`
}

func (theme Theme) Paragraph(text string) string {
	return "\n" + `<p style='
font-size: ` + theme.FontSize + `;
color: #333;
line-height: 1.6;
margin-bottom: 15px;
'>` + text + `</p>` + "\n"
}

// Theme method for Divider
func (theme Theme) Divider() string {
	return `<hr style='
border: 0;
height: 1px;
background: ` + theme.PrimaryColor + `;
margin: 20px 0;
' />`
}

// Theme method for Container with customizable width, height, and background
func (theme Theme) Container(content, width, height, background string) string {
	// Set default values if not provided
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "auto"
	}
	if background == "" {
		background = "transparent"
	}
	return `<div style='
width: ` + width + `;
height: ` + height + `;
margin: ` + theme.DefaultMargin + `;
padding: ` + theme.DefaultPadding + `;
background-color: ` + background + `;
'>` + content + `</div>`
}

// Flexbox alignment for content with customizable horizontal and vertical alignment
func (theme Theme) Align(content, align string) string {
	// Set default alignment to left if not provided
	if align == "" {
		align = "left"
	}
	return `<div style='text-align: ` + align + `;'>` + content + `</div>`
}

// Theme method for Form with action URL, input fields, and submit button
func (theme Theme) Form(fields []string, actionURL string, realmFunc string, submitText string) string {
	formTitle := `<h2 style='
font-size: 24px;
color: ` + theme.PrimaryColor + `;
margin-bottom: ` + theme.DefaultMargin + `;
text-align: center;
'>` + realmFunc + `</h2>`
	return formTitle + `<form action='` + actionURL + `' method='GET' style='
border-radius: ` + theme.BorderRadius + `;
'>
<input type='hidden' name='help'/>
<input type='hidden' name='__func' value='` + realmFunc + `'/>
` + strings.Join(fields, "\n") + `
<button type='submit' style='
background: linear-gradient(to right, ` + theme.PrimaryColor + `, ` + theme.SecondaryColor + `);
border: none;
color: white;
padding: ` + theme.DefaultPadding + `;
font-size: ` + theme.FontSize + `;
cursor: pointer;
border-radius: ` + theme.BorderRadius + `;
margin-top: ` + theme.DefaultMargin + `;
'>
` + submitText + `
</button>
</form>`
}

// Standalone functions for default theme components
func Form(fields []string, actionURL string, realmFunc string, submitText string) string {
	var theme Theme = DefaultTheme()
	return theme.Form(fields, actionURL, realmFunc, submitText)
}

func Navbar(links map[string]string) string {
	var theme Theme = DefaultTheme()
	return theme.Navbar(links)
}

func Button(text, href string) string {
	var theme Theme = DefaultTheme()
	return theme.Button(text, href)
}

func Image(src, width, height string) string {
	var theme Theme = DefaultTheme()
	return theme.Image(src, width, height)
}

func Card(content, width, height string) string {
	var theme Theme = DefaultTheme()
	return theme.Card(content, width, height)
}

func TextInput(name, placeholder string) string {
	var theme Theme = DefaultTheme()
	return theme.TextInput(name, placeholder)
}

func CodeBlock(code string) string {
	var theme Theme = DefaultTheme()
	return theme.CodeBlock(code)
}

func Grid(items []string, columns int, gap string, textAlign string) string {
	var theme Theme = DefaultTheme()
	return theme.Grid(items, columns, gap, textAlign)
}

func Footer(links []string) string {
	var theme Theme = DefaultTheme()
	return theme.Footer(links)
}

func Heading(level int, text string) string {
	var theme Theme = DefaultTheme()
	return theme.Heading(level, text)
}

func Paragraph(text string) string {
	var theme Theme = DefaultTheme()
	return theme.Paragraph(text)
}

func Divider() string {
	var theme Theme = DefaultTheme()
	return theme.Divider()
}

// Applies the default margin to content
func Margin(content, margin string) string {
	var theme Theme = DefaultTheme()
	return theme.Margin(content, margin)
}

// Applies the default padding to content
func Padding(content, margin string) string {
	var theme Theme = DefaultTheme()
	return theme.Padding(content, margin)
}

func Container(content, width, height, background string) string {
	var theme Theme = DefaultTheme()
	return theme.Container(content, width, height, background)
}

func Align(content, align string) string {
	var theme Theme = DefaultTheme()
	return theme.Align(content, align)
}
