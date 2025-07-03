package benchops

func Init(filepath string) {
	initExporter(filepath)
	InitMeasure()
}
