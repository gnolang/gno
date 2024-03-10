package gnolang

import "testing"

func TestTerminatingStatements(t *testing.T) {
	t.Parallel()
	m := NewMachine("test", nil)
	c := `package test
func main() {
	z := y()
	a := s()
	values := []int{0, 1, 2, 3, 4}
	processValues(values)
	validLabel()
	switchLabel()
	b := compare(5,3)
	c := isEqual(3,3)
	fruitStall()
	whichDay()
	
	
	invalidCompareBool(2,3)
	invalidIfStatement(6)
	invalidIfLogic(0)
	example2(4,3)
	invalidExample1(1)
	println(c)
	println(b)
	println(a)
	println(z)
}

func y() int {
	x := 9
	return 9
}

func invalidExample1(x int) {
    if x > 0 {
        println("Positive") 
    } else {
        println("Non-positive") 
    }
}

func add(a, b int) int {
    return a + b
}

func example2(x, y int) {
    if x > y {
        result := add(x, y)
        println("The sum is:", result)
    } else {
        result := add(y, x)
        println("The sum is:", result)
    }
    println("Execution continues...")
}

func isEqual(a, b int) bool {
	if a == b {
		return true
	}
	return false
}


func s() int {
	x := 9
	switch x {
	case 1:
		println(1)
	case 9:
		return x
	default:
		break
	}

	return x
}

func compare(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}



func whichDay() {
	dayOfWeek := 3

	switch dayOfWeek {

	case 1:
		println("Sunday")

	case 2:
		println("Monday")

	case 3:
		println("Tuesday")

	case 4:
		println("Wednesday")

	case 5:
		println("Thursday")

	case 6:
		println("Friday")

	case 7:
		println("Saturday")

	default:
		println("Invalid day")
	}
}

func fruitStall() {
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
				println(v)
			} else {
				println(50)
			}
		case "banana":
			if v > 10 {
				println(v)
			} else {
				println(10)
			}
		case "orange":
			if v > 30 {
				println(v)
			} else {
				println(30)
			}
		case "mango":
			if v == 5 {
				println(v)
			} else {
				println(5)
			}
		case "watermelon":
			if v == 15 {
				println(v)
			} else {
				println(15)
			}
		default:
			println("Did not found any")
		}
	}
}

func processValues(values []int) {
OuterLoop:
	for i := range values {
		switch values[i] {
		case 0:
			println("Zero")
		case 1:
			println("One")
		case 2:
			println("Two")
		case 3:
			println("Three")
			break OuterLoop 
		default:
			println("Other")
		}
	}
}





func validLabel(){
println("validLabel")
OuterLoop:
    for i := 0; i < 10; i++ {
        for j := 0; j < 10; j++ {
            println("i =", i, "j =", j)
            break OuterLoop
        }
    }
}

func switchLabel(){
SwitchStatement:
    switch 1 {
    case 1:
        println(1)
        for i := 0; i < 10; i++ {
            break SwitchStatement
        }
        println(2)
    }
    println(3)
}

func invalidCompareBool(a, b int) bool {
	if a > b {
		return true
	} else {
	}
}

func invalidIfStatement(x int) {
    if x > 5 {
        
    } else {
        println("x is less than or equal to 5")
    }
}

func invalidIfLogic(x int) bool {
  if x > 0 {
    return true  
  } else {
  }
}

`

	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()

}
