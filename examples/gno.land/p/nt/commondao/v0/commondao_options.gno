package commondao

// Option configures the CommonDAO.
type Option func(*CommonDAO)

// WithID assigns a unique identifier to the DAO.
func WithID(id uint64) Option {
	return func(dao *CommonDAO) {
		dao.id = id
	}
}

// WithName assigns a name to the DAO.
func WithName(name string) Option {
	return func(dao *CommonDAO) {
		dao.name = name
	}
}

// WithDescription assigns a description to the DAO.
func WithDescription(description string) Option {
	return func(dao *CommonDAO) {
		dao.description = description
	}
}

// WithAddress assigns a treasury address to the DAO. Hosting realms
// derive it, typically as a realm sub-identity address
// (chain.DerivePkgSubAddr) so each DAO owns a distinct account.
func WithAddress(addr address) Option {
	return func(dao *CommonDAO) {
		dao.addr = addr
	}
}

// WithParent assigns a parent DAO and registers the DAO as one of the
// parent's children, keeping both sides of the tree wired in one step.
func WithParent(p *CommonDAO) Option {
	return func(dao *CommonDAO) {
		dao.parent = p
		if p != nil {
			p.children.Append(dao)
		}
	}
}

// WithCouncilMember assigns a council member to the DAO.
func WithCouncilMember(addr address) Option {
	return func(dao *CommonDAO) {
		dao.council.Add(addr)
	}
}
