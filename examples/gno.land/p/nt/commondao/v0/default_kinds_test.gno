package commondao_test

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/nt/commondao/v0"
)

// New receives a ReadonlyCommonDAO, which exposes no mutators, so a kind
// cannot mutate the host (or its tree) at Propose time. This is enforced
// structurally by the type system, not at runtime: the following method
// does not compile, because ReadonlyCommonDAO has no UpdateCouncil (nor
// any other mutator), and it cannot be downcast to *CommonDAO.
//
//	func (mutatingKind) New(dao commondao.ReadonlyCommonDAO, _ any) (commondao.ProposalDefinition, error) {
//		dao.UpdateCouncil(nil, []address{memberA}) // compile error: undefined
//		return propDef{}, nil
//	}

func TestWithDefaultKinds(cur realm, t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithDefaultKinds(),
	)

	// WithDefaultKinds registers exactly the execution and register-kind kinds.
	uassert.True(t, dao.HasKind("execution"), "expect execution kind registered")
	uassert.True(t, dao.HasKind("register-kind"), "expect register-kind registered")
	uassert.Equal(t, 2, len(dao.KindNames()))

	// A second application collides on the already-registered names.
	uassert.PanicsWithMessage(t, cur, commondao.ErrProposalKindExists.Error(), func() {
		commondao.New(
			commondao.WithDefaultKinds(),
			commondao.WithDefaultKinds(),
		)
	})
}

func TestExecutionKind(cur realm, t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithDefaultKinds(),
	)

	// A nil Fn is rejected at Propose.
	_, err := dao.Propose(memberA, "execution", commondao.ExecutionArgs{Title: "T", Body: "B"})
	uassert.ErrorIs(t, err, commondao.ErrExecutionFuncRequired)

	// Wrong args type is rejected too.
	_, err = dao.Propose(memberA, "execution", "not the args")
	uassert.ErrorIs(t, err, commondao.ErrExecutionFuncRequired)

	// A valid execution proposal runs the supplied Fn on approval.
	ran := false
	p, err := dao.Propose(memberA, "execution", commondao.ExecutionArgs{
		Title: "Do a thing",
		Body:  "details",
		Fn: func(_ int, _ realm) error {
			ran = true
			return nil
		},
	})
	urequire.NoError(t, err)
	uassert.Equal(t, "Do a thing", p.Definition().Title())
	uassert.Equal(t, "details", p.Definition().Body())

	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
	urequire.Equal(t, string(commondao.StatusPassed), string(p.Status()))
	urequire.NoError(t, dao.Execute(p.ID(), cur))
	uassert.True(t, ran, "expect the supplied Fn to run")
	uassert.Equal(t, string(commondao.StatusExecuted), string(p.Status()))
}

func TestRegisterKindKind(cur realm, t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithDefaultKinds(),
	)

	// A nil target DAO is rejected.
	_, err := dao.Propose(memberA, "register-kind", commondao.RegisterKindArgs{Kind: namedKind("new-kind")})
	uassert.ErrorIs(t, err, commondao.ErrRegisterKindDAORequired)

	// A nil / unnamed kind is rejected.
	_, err = dao.Propose(memberA, "register-kind", commondao.RegisterKindArgs{DAO: dao})
	uassert.ErrorIs(t, err, commondao.ErrRegisterKindRequired)
	_, err = dao.Propose(memberA, "register-kind", commondao.RegisterKindArgs{DAO: dao, Kind: namedKind("")})
	uassert.ErrorIs(t, err, commondao.ErrRegisterKindRequired)

	// Registering an already-registered kind is rejected up front.
	_, err = dao.Propose(memberA, "register-kind", commondao.RegisterKindArgs{DAO: dao, Kind: namedKind("execution")})
	uassert.ErrorIs(t, err, commondao.ErrProposalKindExists)

	// A valid register-kind proposal registers the kind on execution.
	uassert.False(t, dao.HasKind("new-kind"))
	p, err := dao.Propose(memberA, "register-kind", commondao.RegisterKindArgs{
		DAO:  dao,
		Kind: namedKind("new-kind"),
	})
	urequire.NoError(t, err)

	urequire.NoError(t, dao.Vote(memberA, p.ID(), commondao.ChoiceYes, ""))
	urequire.Equal(t, string(commondao.StatusPassed), string(p.Status()))
	urequire.NoError(t, dao.Execute(p.ID(), cur))
	uassert.Equal(t, string(commondao.StatusExecuted), string(p.Status()))

	// The newly registered kind is now proposable.
	uassert.True(t, dao.HasKind("new-kind"))
	_, err = dao.Propose(memberA, "new-kind", nil)
	urequire.NoError(t, err)
}

func TestDeregisterKindSelfBrick(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithDefaultKinds(),
	)

	// The reserved register-kind name cannot be deregistered (self-brick guard).
	err := dao.DeregisterKind("register-kind")
	uassert.ErrorIs(t, err, commondao.ErrCannotDeregisterRegisterKind)
	uassert.True(t, dao.HasKind("register-kind"), "expect register-kind to survive deregistration")

	// Non-reserved default kinds can still be deregistered normally.
	urequire.NoError(t, dao.DeregisterKind("execution"))
	uassert.False(t, dao.HasKind("execution"))
}
