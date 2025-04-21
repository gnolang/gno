package gnolang

type Stage string

const (
	StagePre Stage = "StagePre" // e.g. preprocessing TODO use
	StageAdd Stage = "StageAdd" // e.g. init()
	StageRun Stage = "StageRun" // e.g. main()
)
