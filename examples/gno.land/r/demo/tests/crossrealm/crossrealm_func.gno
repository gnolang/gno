package crossrealm

type F func() bool

var f F

func SetCallback(ff F) {
	f = ff
}

func ExecuteCallback() {
	f()
}
