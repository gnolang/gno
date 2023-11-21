package controller

type Controller interface {
	Next(input, setPoint, rawOutput float64) float64
}