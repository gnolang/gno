package std

type Banker interface {
	B()
}

func TVParam(m struct{ m1 map[string]any })

func TVResult() interface{}

func TVFull(map[Banker]map[string]any) (n [86]map[string]bool)
