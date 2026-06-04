package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticAnalysisShouldPanic(t *testing.T) {
	cases := []struct {
		name string
		code string
	}{
		{
			name: "Test Case 1",
			code: `package test
				func main() {
					invalidSwitch()
				}
				func invalidSwitch() int {
					mockVar := map[string]int{
						"apple":      10,
						"banana":     20,
						"orange":     30,
						"mango":      5,
						"watermelon": 15,
					}

					for k, v := range mockVar {
						switch k {
						case "apple":
							if v < 50 {
								return v
							} else {
								return 50
							}
						case "banana":
							if v > 10 {
								return v
							} else {
								return 10
							}
						case "orange":
							if v > 30 {
							} else {
								return 30
							}
						case "mango":
							if v == 5 {
								return v
							} else {
								return 5
							}
						case "watermelon":
							if v == 15 {
								return v
							} else {
								return 15
							}
						default:
							return 0
						}
					}
				}
		`,
		},
		{
			name: "Test Case 2",
			code: `package test
				func main() {
		
				}
		
				func invalidLabel() int{
					Outer:
						for {
							for {
								break Outer
							}
						}
				}
		`,
		},
		{
			name: "Test Case 3",
			code: `package test
				func main() {
					oddOrEven(3)
				}
		
				func oddOrEven(x int) bool {
					for i := 0; i < x; i++ {
						if x%2 == 0 {
		
						} else {
							return false
						}
					}
				}
		`,
		},
		{
			name: "Test Case 4",
			code: `package test
				func main() {
					sumArr([]int{1,1,1})
				}
		
				func sumArr(x []int) int {
					sum := 0
					for _, s := range x {
						sum = sum + s
					}
				}
		`,
		},
		{
			name: "Test Case 5",
			code: `package test
				func main() {
					invalidSwitch(3)
				}
		
				func invalidSwitch(x int) int {
					switch x {
						case 1:
							return 1
						case 2:
							return 2
					}
				}
		`,
		},
		{
			name: "Test Case 6",
			code: `package test
				func main() {
					invalidCompareBool(4,5)
				}
		
				func invalidCompareBool(a, b int) bool {
					if a > b {
						return true
					} else {
					}
				}
		`,
		},
		{
			name: "Test Case 7",
			code: `package test
				func main() {
					invalidIfStatement(6)
				}
		
				func invalidIfStatement(x int) int {
				   if x > 5 {
		
				   } else {
				       return x
				   }
				}
		`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testFunc := func() {
				m := NewMachine("test", nil)

				n := m.MustParseFile("main.go", tc.code)
				m.RunFiles(n)
				m.RunMain()
			}

			assert.Panics(t, testFunc, "The code did not panic")
		})
	}
}

func TestStaticAnalysisShouldPass(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			name: "Test Case 1",
			code: `package test
			func main() {
				first := a()
			}
		
			func a() int {
				x := 9
				return 9
			}
		`,
		},
		{
			name: "Test Case 2",
			code: `package test
			func main() {
			validLabel()
		}
		
		func validLabel() int{
		OuterLoop:
		for i := 0; i < 10; i++ {
		    for j := 0; j < 10; j++ {
		        break OuterLoop
		    }
		}
			return 0
		}
		`,
		},
		{
			name: "Test Case 3",
			code: `package test
		func main() {
			fruitStall()
		
		}
		
		func fruitStall() int {
			mockVar := map[string]int{
				"apple":      10,
				"banana":     20,
				"orange":     30,
				"mango":      5,
				"watermelon": 15,
			}
		
			for k, v := range mockVar {
				switch k {
				case "apple":
					if v < 50 {
						return(v)
					} else {
						return(50)
					}
				case "banana":
					if v > 10 {
						return(v)
					} else {
						return(10)
					}
				case "orange":
					if v > 30 {
						return(v)
					} else {
						return(30)
					}
				case "mango":
					if v == 5 {
						return(v)
					} else {
						return(5)
					}
				case "watermelon":
					if v == 15 {
						return(v)
					} else {
						return(15)
					}
				default:
					return 0
				}
			}
			return 0
		}
		`,
		},
		{
			name: "Test Case 4",
			code: `
			package test
			func main() {
				whichDay()
			}
		func whichDay() string{
			dayOfWeek := 3
		
			switch dayOfWeek {
		
			case 1:
				return "Sunday"
		
			case 2:
				return "Monday"
		
			case 3:
				return "Tuesday"
		
			case 4:
				return "Wednesday"
		
			case 5:
				return "Thursday"
		
			case 6:
				return "Friday"
		
			case 7:
				return "Saturday"
		
			default:
				return "Invalid day"
			}
			return "Not a day"
		}
		
		`,
		},
		{
			name: "Test Case 5",
			code: `package test
			func main() {
			switchLabel()
		}
		
		func switchLabel() int{
		SwitchStatement:
		switch 1 {
		case 1:
		    return 1
		    for i := 0; i < 10; i++ {
		        break SwitchStatement
		    }
		    return 2
		}
		return 3
		}
		`,
		},
		{
			name: "Test Case 6",
			code: `
			package test
			func main() {
				add(1,1)
			}
		func add(a, b int) int {
		return a + b
		}`,
		},
		{
			name: "Test Case 7",
			code: `package test
				func main() {
					z := y()
				}
		
			func y() int{
				x := 9
				return 9
			}
		`,
		},
		{
			name: "Test Case 8",
			code: `package test
				func main() {
					isEqual(2,2)
				}
				func isEqual(a, b int) bool {
				if a == b {
				return true
			}
				return false
			}
		`,
		},
		{
			name: "Test Case 9",
			code: `
		package test
				func main() {
					f(2)	
				}
		func f(a int) int {
			switch a {
			case 1:
				return 1
			default:
				return 0
			}
		} 
		`,
		},
		{
			name: "Test Case 10",
			code: `
		package test
				func main() {
					f(0)	
				}
		func f(a int) int {
			if a > 0 {
				return 1
			} else {
				return 0
			}
		}
		`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMachine("test", nil)
			n := m.MustParseFile("main.go", tc.code)
			m.RunFiles(n)
			m.RunMain()
		})
	}
}
