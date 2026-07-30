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

func TestExecutionKind(cur realm, t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithProposalKind(commondao.ExecutionKind{}),
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

// TestDeregisterKind pins that DeregisterKind is a plain primitive with no
// reserved names: any registered kind can be removed (there is no /p/
// self-brick — anti-brick is now the consuming realm's policy), and an
// unknown name reports ErrProposalKindNotFound.
func TestDeregisterKind(t *testing.T) {
	dao := commondao.New(
		commondao.WithCouncilMember(memberA),
		commondao.WithProposalKind(commondao.ExecutionKind{}),
		// "register-kind" used to be un-deregisterable; it is now just an
		// ordinary registered kind and must deregister like any other.
		commondao.WithProposalKind(namedKind("register-kind")),
	)

	// Any registered kind deregisters, including the former reserved name.
	urequire.NoError(t, dao.DeregisterKind("register-kind"))
	uassert.False(t, dao.HasKind("register-kind"))

	urequire.NoError(t, dao.DeregisterKind("execution"))
	uassert.False(t, dao.HasKind("execution"))

	// An unknown name is rejected.
	err := dao.DeregisterKind("unknown")
	uassert.ErrorIs(t, err, commondao.ErrProposalKindNotFound)
}
