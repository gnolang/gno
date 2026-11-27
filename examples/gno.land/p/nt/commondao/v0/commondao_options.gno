package commondao

// Option configures the CommonDAO. Options are values with a method, not
// closures: a closure made in a realm's init runs read-only when New calls it.
type Option interface {
	apply(*CommonDAO)
}

type (
	withID            uint64
	withName          string
	withPurpose       string
	withDescription   string
	withAddress       address
	withParent        struct{ p *CommonDAO }
	withCouncilMember address
	withProposalKind  struct{ k ProposalKind }
)

func (o withID) apply(dao *CommonDAO)          { dao.id = uint64(o) }
func (o withName) apply(dao *CommonDAO)        { dao.name = string(o) }
func (o withPurpose) apply(dao *CommonDAO)     { dao.purpose = string(o) }
func (o withDescription) apply(dao *CommonDAO) { dao.description = string(o) }
func (o withAddress) apply(dao *CommonDAO)     { dao.addr = address(o) }

func (o withParent) apply(dao *CommonDAO) {
	dao.parent = o.p
	if o.p != nil {
		o.p.children.Append(dao)
	}
}

func (o withCouncilMember) apply(dao *CommonDAO) { dao.council.Add(address(o)) }

func (o withProposalKind) apply(dao *CommonDAO) {
	if err := dao.RegisterKind(o.k); err != nil {
		panic(err)
	}
}

// WithID assigns a unique identifier to the DAO.
func WithID(id uint64) Option {
	return withID(id)
}

// WithName assigns a name to the DAO.
func WithName(name string) Option {
	return withName(name)
}

// WithPurpose assigns a purpose to the DAO. Purpose and description
// together form the DAO's Charter (docs/CONSTITUTION.md :1485).
func WithPurpose(purpose string) Option {
	return withPurpose(purpose)
}

// WithDescription assigns a description to the DAO.
func WithDescription(description string) Option {
	return withDescription(description)
}

// WithAddress assigns a treasury address to the DAO. Hosting realms
// derive it, typically as a realm sub-identity address
// (chain.DerivePkgSubAddr) so each DAO owns a distinct account.
func WithAddress(addr address) Option {
	return withAddress(addr)
}

// WithParent assigns a parent DAO and registers the DAO as one of the
// parent's children, keeping both sides of the tree wired in one step.
func WithParent(p *CommonDAO) Option {
	return withParent{p: p}
}

// WithCouncilMember assigns a council member to the DAO.
func WithCouncilMember(addr address) Option {
	return withCouncilMember(addr)
}

// WithProposalKind registers a proposal kind on the DAO.
// It panics when the kind is nil, has an empty name, or its name is
// already registered.
//
// The package ships one concrete kind, ExecutionKind; seed it with
// WithProposalKind(ExecutionKind{}).
func WithProposalKind(k ProposalKind) Option {
	return withProposalKind{k: k}
}
