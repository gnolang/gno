package controller_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/gc/pacer/controller"
)

func TestNewPI(t *testing.T) {
	cfg := controller.Config{
		Kp:     1.0,
		Ti:     1.0,
		Tt:    	1.0,
		Period: 1.0,
		Min:    -1,
		Max:    1.0,
	}

	c := controller.NewPI(&cfg)

	if c.Kp != cfg.Kp {
		t.Errorf("Kp not set: %f", c.Kp)
	}

	if c.Ti != cfg.Ti {
		t.Errorf("Ti not set: %f", c.Ti)
	}

	if c.Tt != cfg.Tt {
		t.Errorf("Tt not set: %f", c.Tt)
	}

	if c.Period != cfg.Period {
		t.Errorf("Period not set: %f", c.Period)
	}

	if c.Min != cfg.Min {
		t.Errorf("Min not set: %f", c.Min)
	}

	if c.Max != cfg.Max {
		t.Errorf("Max not set: %f", c.Max)
	}
}

func TestNext(t *testing.T) {
	cfg := controller.Config{
		Kp:     1.0,
		Ti:     1.0,
		Tt:    	1.0,
		Period: 1.0,
		Min:    -1,
		Max:    1.0,
	}

	c := controller.PIController {
		Config: cfg,
		Integral: 0.0,
	}

	input := 0.5
	setPoint := 1.0
	rawOutput := 0.0

	output := c.Next(input, setPoint, rawOutput)

	if output < c.Min || output > c.Max {
		t.Errorf("Output out of bounds: %f", output)
	}
}

func TestUpdate(t *testing.T) {
    cfg := controller.Config{
		Kp:     1.0,
		Ti:     1.0,
		Tt:    	1.0,
		Period: 1.0,
		Min:    -1,
		Max:    1.0,
	}

	c := controller.PIController {
		Config: cfg,
		Integral: 0.0,
	}

    input := 0.5
    setPoint := 1.0
    rawOutput := 0.5
    output := 0.5

    c.Update(input, setPoint, rawOutput, output)

	if c.Integral != 0.5 {
		t.Errorf("Integral not updated: %f", c.Integral)
	}
}

func TestOutputAdjustToMin(t *testing.T) {
	cfg := controller.Config{
		Kp:     1.0,
		Ti:     1.0,
		Tt:    	1.0,
		Period: 1.0,
		Min:    1.0,
		Max:    2.0,
	}

	c := controller.PIController {
		Config: cfg,
		Integral: 0.0,
	}

	input := 0.5
	setPoint := 1.0

	_, output := c.Output(input, setPoint)
	fmt.Println(output)
	if output != cfg.Min {
		t.Errorf("Output not adjusted: %f", output)
	}
}

func TestOutputAdjustToMax(t *testing.T) {
	cfg := controller.Config{
		Kp:     1.0,
		Ti:     1.0,
		Tt:    	1.0,
		Period: 1.0,
		Min:    -2.0,
		Max:    -1.5,
	}

	c := controller.PIController {
		Config: cfg,
		Integral: 0.0,
	}

	input := 0.5
	setPoint := 1.0

	_, output := c.Output(input, setPoint)
	if output != cfg.Max {
		t.Errorf("Output not adjusted: %f", output)
	}
}