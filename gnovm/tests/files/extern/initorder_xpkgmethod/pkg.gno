package initorder_xpkgmethod

// B is a package-level variable. GetB references it so that any caller of
// GetB transitively depends on B. A is declared after GetB here to stress
// that the intra-package init order is also correct (B before A).

var B = 42

type T struct{}

func (T) GetB() int { return B }

var A = T{}.GetB()
