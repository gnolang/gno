package controller

// Proportional Integral Controller
type PIController struct {
	Config
	integral float64
}

type Config struct {
	Kp 		float64	 `json:"k_p"`
	Ti 		float64	 `json:"t_i"`
	Tt 		float64	 `json:"t_t"`
	Period 	float64	 `json:"period"`
	Min 	float64	 `json:"min"`
	Max 	float64	 `json:"max"`
}

func NewPI(cfg *Config) *PIController {
	return &PIController{Config: *cfg}
}

func (c *PIController) output(input, setPoint float64) (rawOutput, output float64) {
	prop := c.Kp * (setPoint - input)
	rawOutput = prop + c.integral
	output = rawOutput

	if output < c.Min {
		output = c.Min
	} else if output > c.Max {
		output = c.Max
	}

	return rawOutput, output
}

func (c *PIController) update(input, setPoint, rawOutput, output float64) {
	if c.Ti != 0 && c.Tt != 0 {
		c.integral += (c.Kp * c.Period / c.Ti) * (setPoint - input) + (c.Period / c.Tt) * (output - rawOutput)
	}
}

func (c *PIController) Next(input, setPoint, rawOutput float64) float64 {
	rawOutput, output := c.output(input, setPoint)
	c.update(input, setPoint, rawOutput, output)
	return output
}