package t75sub

type Obj struct{ N int }

func (o *Obj) Mint(s string, rlm realm) string { return rlm.Sub(s).PkgPath() }

func Mint(s string, rlm realm) string { return rlm.Sub(s).PkgPath() }
