package gas

// OutOfGasError defines an error thrown when an action results in out of gas.
type OutOfGasError struct {
	Descriptor string
}

func (oog OutOfGasError) Error() string {
	return "out of gas in location: " + oog.Descriptor
}

// OverflowError defines an error thrown when an action results in an
// integer/float overflow in gas calculation.
type OverflowError struct {
	Descriptor string
}

func (og OverflowError) Error() string {
	return "gas overflow in location: " + og.Descriptor
}

// PrecisionError defines an error thrown when an action results in a
// precision loss in gas calculation.
type PrecisionError struct {
	Descriptor string
}

func (pe PrecisionError) Error() string {
	return "gas precision loss in location: " + pe.Descriptor
}
