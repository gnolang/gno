package gnolang

import "testing"

func TestStaticAnalysisShouldPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	cases := []string{
		`
	package test 
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
}
`,
		`package test
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
		`package test
		func main() { 
			oddOrEven(3)
		}

	func oddOrEven(x int) bool {
		for i := 0; i < x; i++ {
			if(x % 2 == 0){
			
			} else {
				return false 
			}
		}
	}
	`,
		`package test
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
		`package test
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
		`package test
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
		`package test
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
	}

	for _, s := range cases {
		m := NewMachine("test", nil)

		n := MustParseFile("main.go", s)
		m.RunFiles(n)
		m.RunMain()
	}
}

func TestStaticAnalysisShouldPass(t *testing.T) {
	cases := []string{
		`package test
func main() {
	first := a()
	println(first)
}

func a() int {
	x := 9
	return 9
}
`, `package test
	func main() {
	validLabel()	
}

func validLabel() int{
println("validLabel")
OuterLoop:
    for i := 0; i < 10; i++ {
        for j := 0; j < 10; j++ {
            println("i =", i, "j =", j)
            break OuterLoop
        }
    }
	return 0
}
`, `package test
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
`, `
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

`, `package test
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
`, `
	package test 
	func main() {
		add(1,1)	
	}
func add(a, b int) int {
    return a + b
}`,
		`package test
		func main() { 
			z := y()
		}

	func y() int{
		x := 9
		return 9
	}
`, `package test
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
	}

	for _, s := range cases {
		m := NewMachine("test", nil)

		n := MustParseFile("main.go", s)
		m.RunFiles(n)
		m.RunMain()
	}
}
