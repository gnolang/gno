package gnoffee

import (
	"testing"
)

func TestNormalizeGoCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Basic normalization",
			input: `
				func main() {
					println("Hello, World!")
				}
			`,
			expected: `func main() {
	println("Hello, World!")
}`,
		},
		{
			name: "No indentation",
			input: `func main() {
println("Hello, World!")
}`,
			expected: `func main() {
println("Hello, World!")
}`,
		},
		{
			name: "Mixed indentation 1",
			input: `
				func main() {
			  println("Hello, World!")
			}`,
			expected: `func main() {
  println("Hello, World!")
}`,
		},
		{
			name: "Mixed indentation 2",
			input: `
			func main() {
			  println("Hello, World!")
				}`,
			expected: `func main() {
  println("Hello, World!")
	}`,
		},
		{
			name:     "Only one line with spaces",
			input:    "       single line with spaces",
			expected: "single line with spaces",
		},
		{
			name: "Empty lines",
			input: `

				func main() {

					println("Hello!")

				}

			`,
			expected: `func main() {

	println("Hello!")

}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			normalized := normalizeGoCode(test.input)
			if normalized != test.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", test.expected, normalized)
			}
		})
	}
}
