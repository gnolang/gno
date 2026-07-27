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

// WithParent assigns a parent DAO.
func WithParent(p *CommonDAO) Option {
	return func(dao *CommonDAO) {
		dao.parent = p
	}
}

// WithCouncilMember assigns a council member to the DAO.
func WithCouncilMember(addr address) Option {
	return func(dao *CommonDAO) {
		dao.council.Add(addr)
	}
}

// WithMaxActiveProposals sets the cap for simultaneously active proposals.
// Values below one are ignored: a DAO must always be able to propose.
func WithMaxActiveProposals(max int) Option {
	return func(dao *CommonDAO) {
		if max >= 1 {
			dao.maxActiveProposals = max
		}
	}
}
