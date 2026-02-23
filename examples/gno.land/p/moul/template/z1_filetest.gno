package main

import "gno.land/p/moul/template"

func main() {
	markdownTemplate := `{{ H1 "Welcome to {{ .Name }}'s Page!" }}

{{ Bold "Skills" -}}
{{- range .Skills }}
{{ BulletItem . -}}
{{- end }}

{{ if .ShowProjects -}}
{{ H2 "Projects" }}
{{- range .Projects }}
{{ H3 .Name }}
{{ Paragraph .Description }}
{{- if .HasDemo }}
{{ Link "View Demo" .DemoURL }}
{{- end }}
{{- end }}
{{- end }}

{{ H2 "Contact Information" }}
{{- range .Contacts }}
{{ BulletItem (printf "%s: %s" .Platform .Value) -}}
{{- end }}

{{ H2 "Recent Posts" }}
{{- if .HasPosts }}
{{- range .RecentPosts }}
{{ H3 .Title }}
{{ Italic .Date }}
{{ Paragraph .Summary }}
{{- if .HasComments }}
{{ Blockquote (printf "Latest comment: %s" (index .Comments 0)) }}
{{- end }}
{{ HorizontalRule -}}
{{- end }}
{{- end }}`

	data := map[string]interface{}{
		"Name": "Alice",
		"Skills": []interface{}{
			"Go Programming",
			"Smart Contracts",
			"Blockchain Development",
			"System Architecture",
		},
		"ShowProjects": true,
		"Projects": []interface{}{
			map[string]interface{}{
				"Name":        "Gno Templates",
				"Description": "A template engine for Gno",
				"HasDemo":     true,
				"DemoURL":     "https://example.com/demo",
			},
			map[string]interface{}{
				"Name":        "Blockchain Explorer",
				"Description": "A tool for exploring the Gno chain",
				"HasDemo":     false,
			},
		},
		"Contacts": []interface{}{
			map[string]interface{}{
				"Platform": "Email",
				"Value":    "alice@gno.land",
			},
			map[string]interface{}{
				"Platform": "GitHub",
				"Value":    "@alice",
			},
		},
		"HasPosts": true,
		"RecentPosts": []interface{}{
			map[string]interface{}{
				"Title":       "Getting Started with Gno",
				"Date":        "2024-03-20",
				"Summary":     "A beginner's guide to Gno development",
				"HasComments": true,
				"Comments":    []interface{}{"Great article!", "Very helpful"},
			},
			map[string]interface{}{
				"Title":       "Advanced Template Techniques",
				"Date":        "2024-03-19",
				"Summary":     "Deep dive into Gno templates",
				"HasComments": false,
			},
		},
	}

	renderer := template.NewRenderer()
	output := renderer.Render(markdownTemplate, data)
	println(output)
}

// Output:
// # Welcome to Alice's Page!
//
// **Skills**
// - Go Programming
//
// - Smart Contracts
//
// - Blockchain Development
//
// - System Architecture
//
//
//
// ## Projects
// ### Gno Templates
// A template engine for Gno
// [View Demo](https://example.com/demo)
//
//
// ### Blockchain Explorer
// A tool for exploring the Gno chain
//
// ## Contact Information
// - Email: alice@gno.land
//
// - GitHub: @alice
//
//
// ## Recent Posts
// ### Getting Started with Gno
// *2024-03-20*
// A beginner's guide to Gno development
// > Latest comment: Great article!
//
// ---
//
// ### Advanced Template Techniques
// *2024-03-19*
// Deep dive into Gno templates
// ---
