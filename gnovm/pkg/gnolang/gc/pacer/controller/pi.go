package controller

// Proportional Integral Controller
type PIController struct {
	Config
	Integral float64
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

func (c *PIController) Output(input, setPoint float64) (rawOutput, output float64) {
	prop := c.Kp * (setPoint - input)
	rawOutput = prop + c.Integral
	output = rawOutput

	if output < c.Min {
		output = c.Min
	} else if output > c.Max {
		output = c.Max
	}

	return rawOutput, output
}

func (c *PIController) Update(input, setPoint, rawOutput, output float64) {
	if c.Ti != 0 && c.Tt != 0 {
		c.Integral += (c.Kp * c.Period / c.Ti) * (setPoint - input) + (c.Period / c.Tt) * (output - rawOutput)
	}
}

func (c *PIController) Next(input, setPoint, rawOutput float64) float64 {
	rawOutput, output := c.Output(input, setPoint)
	c.Update(input, setPoint, rawOutput, output)
	return output
}