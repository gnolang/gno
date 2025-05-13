package gnolang

type Stage string

const (
	StagePre Stage = "StagePre" // e.g. static evaluation during preprocessing
	StageAdd Stage = "StageAdd" // e.g. init()
	StageRun Stage = "StageRun" // e.g. main()
)
