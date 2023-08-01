package governor

import "gno.land/p/demo/governance/governor"
import "fmt"

func main() {
	vvf := governor.NewGovernorVotesVetoFraction(33)
	fmt.Printf(" %0.2f %% \n", vvf.VetoPercentage())
	println("done")
	var a float64
	a = 100
	// println(100 * vvf.VetoPercentage())
	fmt.Printf(" %0.2f \n", a*vvf.VetoPercentage())
}

// Output:
// 0.33 %
// done
//  32.67
