package evaluation

import (
	"std"
	"testing"
)

func TestEvaluationCommityMembers(t *testing.T) {
	members := []std.Address{"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", "g1qpymzwx4l4cy6cerdyajp9ksvjsf20rk5y9rtt", "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"}
	ec := NewEvalutaionCommity()

	t.Run("Designate Evaluation Commity Members", func(t *testing.T) {
		ec.designateMembers(members)
		if !equal(ec.members, members) {
			t.Errorf("Designated Evaluation Commity members got %v expected %v", members, ec.members)
		}
	})

	t.Run("Dismiss Evaluation Commity Members", func(t *testing.T) {
		ec.dismissMembers(members)
	})
}
