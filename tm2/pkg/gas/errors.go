package gas

// OutOfGasError defines an error thrown when an action results in out of gas.
type OutOfGasError struct {
	Descriptor string
}

func (oog OutOfGasError) Error() string {
	return "out of gas in location: " + oog.Descriptor
}

// OverflowError defines an error thrown when an action results gas consumption
// unsigned integer overflow.
type OverflowError struct {
	Descriptor string
}

func (oog OverflowError) Error() string {
	return "gas overflow in location: " + oog.Descriptor
}
