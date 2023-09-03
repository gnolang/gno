package gnoffee

import (
	"testing"
)

func TestStage1(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Basic Export Functionality",
			input: `
export Foo as FooInstance
invar BarInterface = Baz
`,
			expected: `
//gnoffee:export Foo as FooInstance
//gnoffee:invar BarInterface
var BarInterface = Baz
`,
		},
		{
			name: "Complex Input with Mixed Code",
			input: `
func someFunction() {
	println("Hello, World!")
}

export Baz as BazInstance
invar QuxInterface = Baz

func anotherFunction() bool {
	return true
}

export Quux as QuuxInstance
`,
			expected: `
func someFunction() {
	println("Hello, World!")
}

//gnoffee:export Baz as BazInstance
//gnoffee:invar QuxInterface
var QuxInterface = Baz

func anotherFunction() bool {
	return true
}

//gnoffee:export Quux as QuuxInstance
`,
		},
		{
			name: "Input with No Changes",
			input: `
func simpleFunction() {
	println("Just a simple function!")
}
`,
			expected: `
func simpleFunction() {
	println("Just a simple function!")
}
`,
		},
		{
			name: "Already Annotated Source",
			input: `
// Some comment
//gnoffee:export AlreadyExported as AlreadyInstance
func someFunction() {
    println("This function is already annotated!")
}

//gnoffee:invar AlreadyInterface
var AlreadyInterface Already
`,
			expected: `
// Some comment
//gnoffee:export AlreadyExported as AlreadyInstance
func someFunction() {
    println("This function is already annotated!")
}

//gnoffee:invar AlreadyInterface
var AlreadyInterface Already
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := Stage1(tt.input)

			if output != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s\n", tt.expected, output)
			}
		})
	}
}
