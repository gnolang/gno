package std

type Banker interface {
	B()
}

func TVParam(m struct{ m1 map[string]interface{} })

func TVResult() interface{ S() map[int]Banker }

func TVFull(map[Banker]map[string]interface{}) (n [86]map[string]bool)
