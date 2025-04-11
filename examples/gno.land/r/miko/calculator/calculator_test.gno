package calculator

import "testing"

func TestCounter_Addition(t *testing.T) {

	// Increment the value
	value := ComputeResult("1p1")

	// Verify the value is equal to 2
	if value != "2" {
		t.Fatalf("1 + 1 is not equal to 2")
	}
}

func TestCounter_Subtraction(t *testing.T) {

	// Increment the value
	value := ComputeResult("1-1")

	// Verify the value is equal to 0
	if value != "0" {
		t.Fatalf("1 - 1 is not equal to 0")
	}
}

func TestCounter_Multiplication(t *testing.T) {

	// Increment the value
	value := ComputeResult("1*4")

	// Verify the value is equal to 4
	if value != "4" {
		t.Fatalf("1 * 4 is not equal to 4")
	}
}

func TestCounter_Division(t *testing.T) {

	// Increment the value
	value := ComputeResult("4/2")

	// Verify the value is equal to 2
	if value != "2" {
		t.Fatalf("4 / 2 is not equal to 2")
	}
}

func TestCounter_AdditionDecimal(t *testing.T) {

	// Increment the value
	value := ComputeResult("1.2p1.3")

	// Verify the value is equal to 2.5
	if value != "2.5" {
		t.Fatalf("1.2 + 1.3 is not equal to 2.5")
	}
}

func TestCounter_SubtractionDecimal(t *testing.T) {

	// Increment the value
	value := ComputeResult("1.3-1.2")

	// Verify the value is equal to 0.1
	if value != "0.1" {
		t.Fatalf("1.3 - 1.2 is not equal to 0.1")
	}
}

func TestCounter_MultiplicationDecimal(t *testing.T) {

	// Increment the value
	value := ComputeResult("3*1.5")

	// Verify the value is equal to 4.5
	if value != "4.5" {
		t.Fatalf("3 * 1.5 is not equal to 4.5")
	}
}

func TestCounter_DivisionDecimal(t *testing.T) {

	// Increment the value
	value := ComputeResult("2/0.5")

	// Verify the value is equal to 4
	if value != "4" {
		t.Fatalf("2 / 0.5 is not equal to 4")
	}
}
