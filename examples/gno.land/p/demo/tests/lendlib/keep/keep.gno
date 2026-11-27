package keep

// Holder keeps a function for later.
type Holder struct{ F func() }

func Keep(h *Holder, f func()) { h.F = f }

func (h *Holder) Fire() { h.F() }
