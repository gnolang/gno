package gnoimports

type MockResolver struct {
	pkgspath map[string]*Package   // pkg path -> pkg
	pkgs     map[string][]*Package // pkg name -> []pkg
}

func NewMockResolver() *MockResolver {
	return &MockResolver{
		pkgspath: make(map[string]*Package),
		pkgs:     make(map[string][]*Package),
	}
}

func (m *MockResolver) AddPackage(pkg *Package) []*Package {
	m.pkgs[pkg.Name] = append(m.pkgs[pkg.Name], pkg)
	m.pkgspath[pkg.Path] = pkg
	return nil
}

func (m *MockResolver) ResolveName(pkgname string) []*Package {
	return m.pkgs[pkgname]
}

func (m *MockResolver) ResolvePath(pkgpath string) *Package {
	return m.pkgspath[pkgpath]
}
