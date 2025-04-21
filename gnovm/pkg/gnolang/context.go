package gnolang

type ExecKind string

const (
	ExecKindPre ExecKind = "ExecKindPre" // e.g. preprocessing
	ExecKindAdd ExecKind = "ExecKindAdd" // e.g. init()
	ExecKindRun ExecKind = "ExecKindRun" // e.g. main()
)

type ExecContextI interface {
	SetExecKind(ExecKind)
	GetExecKind() ExecKind
}
