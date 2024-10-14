package uif

import (
	"strings"

	"gno.land/p/demo/uif"
)

// RenderHomePage - Renders the Home page with the introduction and navigation
func RenderHomePage() string {
	// Define a custom theme for the page
	customTheme := uif.Theme{
		PrimaryColor:   "#3498db",   // Blue color
		SecondaryColor: "#2ecc71",   // Green color
		FontSize:       "18px",      // Slightly larger font size
		BorderRadius:   "8px",       // More rounded corners
		DefaultPadding: "12px 24px", // Custom padding for better spacing
		DefaultMargin:  "10px",
	}

	// Navbar with links and href values using the theme method
	navbarLinks := map[string]string{
		"Home":       "/r/demo/uif",
		"Components": "/r/demo/uif:components",
		"Themes":     "/r/demo/uif:themes",
		"UI Package": "/r/demo/uif:ui-package",
	}
	navbar := customTheme.Navbar(navbarLinks)

	// Heading for the introduction
	headingIntro := customTheme.Heading(1, "Welcome to the UIF Package Home Page")

	// Introduction paragraph
	paragraphIntro := customTheme.Paragraph(`
		The UIF package offers an easy way to create reusable UI components with a built-in theming system.
		You can either use the default theme or define your own custom theme for a consistent and flexible UI design.
		In this guide, you'll learn how to use the package, see examples, and discover how it can simplify UI development.
		Whether you're creating forms, buttons, or entire layouts, UIF provides the tools you need to build
		a clean, responsive interface.
	`)

	// Feature Section: Benefits of Using UIF
	headingFeatures := customTheme.Heading(2, "Why Use UIF?")
	paragraphFeatures := customTheme.Paragraph(`
		UIF provides several advantages for developers looking to streamline UI development:
		<ul>
			<li><strong>Consistency:</strong> Ensure a uniform look and feel across all components with theming.</li>
			<li><strong>Reusability:</strong> Create reusable UI elements that can be used across multiple pages.</li>
			<li><strong>Flexibility:</strong> Customize each component using the default or custom themes for a personalized experience.</li>
			<li><strong>Responsiveness:</strong> All components are designed to be responsive, adapting to different screen sizes.</li>
		</ul>
	`)

	// How it Works Section
	headingHowItWorks := customTheme.Heading(2, "How Does It Work?")
	paragraphHowItWorks := customTheme.Paragraph(`
		UIF makes it easy to build and customize UI components. With just a few lines of code, you can create buttons,
		cards, text inputs, and other essential elements. Themes allow you to customize colors, fonts, padding, and more
		to fit your application's branding or user preferences.
		You can even create custom themes to meet specific design requirements.
	`)

	// Call-to-action button for exploring the guide
	buttonExplore := customTheme.Button("Explore Now", "/r/demo/uif:components")

	// Divider to separate sections
	divider := customTheme.Divider()

	// Footer for the page
	footer := customTheme.Footer([]string{"Privacy Policy", "Contact", "Terms of Service"})

	// Organize content using grid layout
	gridItems := []string{
		headingIntro + paragraphIntro + divider,
		headingHowItWorks + paragraphHowItWorks + buttonExplore + divider,
		headingFeatures + paragraphFeatures,
	}

	// Create a grid with 1 column, 40px gap, and left-aligned text
	contentGrid := customTheme.Grid(gridItems, 1, "0", "left")

	// Combine everything into the home page
	return navbar + contentGrid + footer
}

func RenderComponentsPage() string {
	// Use default theme for the components page

	// Navbar with links and href values using default theme
	navbarLinks := map[string]string{
		"Home":       "/r/demo/uif",
		"Components": "/r/demo/uif:components",
		"Themes":     "/r/demo/uif:themes",
		"UI Package": "/r/demo/uif:ui-package",
	}
	navbar := uif.Navbar(navbarLinks)

	// Main title and introduction using default theme
	headingMain := uif.Heading(1, "Components Overview")
	descriptionMain := uif.Paragraph(`
		This page provides an overview of the various UI components available in the UIF package. Each component is customizable
		through the use of themes, allowing you to modify colors, padding, and other styling properties. Explore how these
		components can be used to create beautiful, responsive user interfaces.
	`)

	// Section: Button Component
	buttonTitle := uif.Heading(2, "Button Component")
	descriptionButton := uif.Paragraph(`
		Buttons are interactive elements that users can click to trigger actions. You can define the button's text,
		link, and appearance using the default theme. Adjust the padding, background colors, and border radius to fit your design.
	`)
	button := uif.Button("Click Me", "#")
	// Define a custom theme for a large button
	customLargeButtonTheme := uif.Theme{
		PrimaryColor:   "#1abc9c",   // Teal color
		SecondaryColor: "#16a085",   // Darker teal
		DefaultPadding: "16px 32px", // Larger padding for a bigger button
		FontSize:       "20px",      // Bigger font size
		BorderRadius:   "12px",
		DefaultMargin:  "10px",
	}
	buttonLarge := customLargeButtonTheme.Button("Learn More", "#")
	buttonCode := uif.CodeBlock(`
button := uif.Button("Click Me", "#")
buttonLarge := customLargeButtonTheme.Button("Learn More", "#")
	`)

	// Section: Text Input Component
	textInputTitle := uif.Heading(2, "Text Input Component")

	descriptionTextInput := uif.Paragraph(`
		TextInput fields allow users to input text. Customize the placeholder text, padding, and border radius using the default theme.
	`)

	textInput := uif.TextInput("name", "Enter your name")
	textInputEmail := uif.Theme{
		PrimaryColor:   "#2980b9",   // Blue color
		DefaultPadding: "10px 20px", // Custom padding for the input
		FontSize:       "16px",      // Regular font size
		BorderRadius:   "8px",
		DefaultMargin:  "10px", // Rounded corners
	}.TextInput("email", "Enter your email")
	textInputCode := uif.CodeBlock(`
textInput := uif.TextInput("name","Enter your name")
textInputEmail := uif.Theme{
	PrimaryColor: "#2980b9",
	Padding: "10px 20px",
	FontSize: "16px",
	BorderRadius: "8px",
}.TextInput("email","Enter your email")
	`)

	formTitle := uif.Heading(2, "Form Component")
	descriptionForm := uif.Paragraph(`
		Forms allow users to input data, which can then be submitted to the help page of realm.
	`)

	// Form fields using TextInput
	fields := []string{
		uif.TextInput("addr", "Enter address"),
	}

	// Create the form with action URL
	form := uif.Form(fields, "/r/demo/users", "GetUserByAddress", "Submit")

	// Section: Card Component with custom width and height
	cardTitle := uif.Heading(2, "Card Component")
	descriptionCard := uif.Paragraph(`
		The Card component is used to group related content inside a bordered, padded container.
		It’s commonly used to display content in a neatly organized box with a consistent style.
		You can use the theme to modify the card's padding, border, and shadow effects, as well as its width and height.
	`)
	cardContent := uif.Paragraph("This is a card component with some content.")
	card := uif.Card(cardContent, "", "") // Card with default dimensions
	cardWithImage := uif.Card(uif.Image("https://letsenhance.io/static/8f5e523ee6b2479e26ecc91b9c25261e/1015f/MainAfter.jpg", "", "200px")+uif.Paragraph("This card contains an image and text."), "400px", "auto")
	cardCode := uif.CodeBlock(`
cardContent := uif.Paragraph("This is a card component with some content.")
card := uif.Card(cardContent, "", "") // Card with default dimensions
// Card with image and custom dimensions
cardWithImage := uif.Card(uif.Image("https://letsenhance.io/static/8f5e523ee6b2479e26ecc91b9c25261e/1015f/MainAfter.jpg", "300px", "200px") + uif.Paragraph("This card contains an image and text."), "400px", "auto")
	`)

	// Section: Divider Component
	dividerTitle := uif.Heading(2, "Divider Component")
	descriptionDivider := uif.Paragraph(`
		Dividers are horizontal lines used to visually separate sections of content. They can be styled using the default theme
		to change their color, thickness, and margins.
	`)
	divider := uif.Divider()
	dividerCode := uif.CodeBlock(`
divider := uif.Divider()
	`)

	// Section: Image Component with custom dimensions
	imageTitle := uif.Heading(2, "Image Component")
	descriptionImage := uif.Paragraph(`
		The Image component allows you to add pictures to your page. You can adjust the border radius, width, and height
		to fit your design. Images can be displayed in cards or standalone.
	`)
	image := uif.Image("https://letsenhance.io/static/8f5e523ee6b2479e26ecc91b9c25261e/1015f/MainAfter.jpg", "300px", "200px")
	imageCode := uif.CodeBlock(`
image := uif.Image("https://letsenhance.io/static/8f5e523ee6b2479e26ecc91b9c25261e/1015f/MainAfter.jpg", "300px", "200px")
	`)

	// Footer
	footer := uif.Footer([]string{"Privacy Policy", "Contact", "Terms of Service"})

	// Combine everything into the components page
	return navbar + headingMain + descriptionMain +
		imageTitle + descriptionImage + image + imageCode +
		cardTitle + descriptionCard + card + cardWithImage + cardCode +
		buttonTitle + descriptionButton + button + buttonLarge + buttonCode +
		textInputTitle + descriptionTextInput + textInput + textInputEmail + textInputCode +
		formTitle + descriptionForm + form +
		dividerTitle + descriptionDivider + divider + dividerCode +
		footer
}

// RenderThemesPage - Renders the page that explains how to apply themes in UIF
func RenderThemesPage(themeColor string) string {
	// Define available themes in a map
	themes := map[string]uif.Theme{
		"purple": {
			PrimaryColor:   "#8e44ad",   // Purple color
			SecondaryColor: "#9b59b6",   // Light purple
			FontSize:       "18px",      // Larger font size
			DefaultPadding: "12px 24px", // Custom padding
			BorderRadius:   "10px",
			DefaultMargin:  "10px", // Rounded corners
		},
		"blue": {
			PrimaryColor:   "#2980b9",   // Blue color
			SecondaryColor: "#3498db",   // Light blue
			FontSize:       "16px",      // Smaller font size
			DefaultPadding: "10px 20px", // Custom padding
			BorderRadius:   "8px",
			DefaultMargin:  "10px", // Slightly rounded corners
		},
		"green": {
			PrimaryColor:   "#27ae60",   // Green color
			SecondaryColor: "#2ecc71",   // Light green
			FontSize:       "20px",      // Larger font size
			DefaultPadding: "14px 28px", // Larger padding
			BorderRadius:   "12px",
			DefaultMargin:  "10px", // More rounded corners
		},
	}

	// Select the theme based on the themeColor argument
	theme, exists := themes[themeColor]
	if !exists {
		// Default to purple if the theme is not recognized
		theme = themes["purple"]
	}

	// Navbar with links and href values
	navbarLinks := map[string]string{
		"Home":       "/r/demo/uif",
		"Components": "/r/demo/uif:components",
		"Themes":     "/r/demo/uif:themes",
		"UI Package": "/r/demo/uif:ui-package",
	}
	navbar := theme.Navbar(navbarLinks)

	// Main themes overview heading and description
	heading := theme.Heading(1, "Themes Overview")
	paragraph := theme.Paragraph(`
		The UIF package allows you to define and apply themes for consistency across your UI components. Themes help you maintain
		a cohesive design by allowing you to set properties such as primary and secondary colors, font size, border radius,
		and padding, which are applied to all UI elements.
	`)

	// Section: Benefits of Using Themes
	benefitsTitle := theme.Heading(2, "Why Use Themes?")
	benefitsText := theme.Paragraph(`
		Using themes ensures a consistent look and feel across your entire application. By defining colors, font sizes,
		and other styles in one place, you make it easier to update your design. It also allows you to quickly switch between
		different styles based on user preferences or branding requirements.
	`)

	// Section: How to Define a Theme
	defineTitle := theme.Heading(2, "How to Define a Theme")
	defineText := theme.Paragraph(`
		A theme in UIF is defined by specifying a primary color, secondary color, font size, padding, and border radius. These
		properties are applied to all UI components such as buttons, cards, text inputs, and more. Here's an example of how to
		define and apply a theme in UIF:
	`)
	defineCode := theme.CodeBlock(`
theme := uif.Theme{
	PrimaryColor:   "#8e44ad",  // Purple
	SecondaryColor: "#9b59b6",  // Light Purple
	FontSize:       "18px",     // Larger font
	DefaultPadding:        "12px 24px", // Custom padding
	BorderRadius:   "10px",     // Rounded corners
}
button := theme.Button("Click Me", "#")
	`)

	// Dynamic button creation for each theme
	themeButtonsTitle := theme.Heading(2, "Apply a Theme")
	themeButtonsText := theme.Paragraph(`
		Click on the buttons below to see how different themes can be applied across your UI components. Each button represents
		a unique theme with different color schemes and styles.
	`)
	buttonPurple := themes["purple"].Button("Apply Purple Theme", "/r/demo/uif:themes/purple")
	buttonBlue := themes["blue"].Button("Apply Blue Theme", "/r/demo/uif:themes/blue")
	buttonGreen := themes["green"].Button("Apply Green Theme", "/r/demo/uif:themes/green")

	// Section: Example UI Components with Selected Theme
	exampleTitle := theme.Heading(2, "Example Components with the Current Theme")
	exampleText := theme.Paragraph(`
		Below are examples of how the selected theme is applied to various UI components. You can see how the colors, padding,
		and other styles change based on the current theme.
	`)
	buttonExample := theme.Button("Sample Button", "#")
	cardExample := theme.Card(theme.Paragraph("This is a sample card with the current theme."), "300px", "auto")
	inputExample := theme.TextInput("name", "Enter your name")

	// Divider
	divider := theme.Divider()

	// Footer
	footer := theme.Footer([]string{"Privacy Policy", "Contact", "Terms of Service"})

	// Combine everything into the themes page
	return navbar +
		heading +
		paragraph +
		benefitsTitle +
		benefitsText +
		defineTitle +
		defineText +
		defineCode +
		themeButtonsTitle +
		themeButtonsText +
		buttonPurple + buttonBlue + buttonGreen +
		exampleTitle +
		exampleText +
		buttonExample + cardExample + inputExample +
		divider + footer
}

// RenderUIPackagePage - Renders the UI Package page
func RenderUIPackagePage() string {
	// Define a theme for the UI Package page
	theme := uif.Theme{
		PrimaryColor:   "#2980b9",   // Dark blue color
		SecondaryColor: "#3498db",   // Light blue color
		FontSize:       "18px",      // Larger font size
		DefaultPadding: "12px 24px", // Custom padding
		BorderRadius:   "10px",
		DefaultMargin:  "10px", // Rounded corners
	}

	// Navbar with links and href values
	navbarLinks := map[string]string{
		"Home":       "/r/demo/uif",
		"Components": "/r/demo/uif:components",
		"Themes":     "/r/demo/uif:themes",
		"UI Package": "/r/demo/uif:ui-package",
	}
	navbar := theme.Navbar(navbarLinks)

	// UI Package page content
	heading := theme.Heading(1, "UI Package Overview")
	paragraph := theme.Paragraph(`The UIF package includes everything you need to build beautiful and responsive user interfaces.`)

	// Button
	button := theme.Button("Get Started with UIF", "#")

	// Divider
	divider := theme.Divider()

	// Footer
	footer := theme.Footer([]string{"Privacy Policy", "Contact", "Terms of Service"})

	// Combine everything into the UI package page
	return navbar + heading + paragraph + button + divider + footer
}

// Render - the main entry point for displaying different pages
func Render(path string) string {
	segments := strings.Split(path, "/")

	if len(segments) > 1 && segments[0] == "themes" {
		themeColor := segments[1]
		return RenderThemesPage(themeColor)
	}

	switch segments[0] {
	case "components":
		return RenderComponentsPage()
	case "themes":
		return RenderThemesPage("")
	case "ui-package":
		return RenderUIPackagePage()
	default:
		// Default to the home page
		return RenderHomePage()
	}
}
