package commondao_test

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/nt/commondao/v0"
)

func TestMemberGroupNew(t *testing.T) {
	g, err := commondao.NewMemberGroup("", nil)
	urequire.ErrorContains(t, err, "member storage is required")

	storage := commondao.NewMemberStorage()
	g, err = commondao.NewMemberGroup("", storage)
	urequire.ErrorContains(t, err, "member group name is required")

	name := "Tier 1"
	g, err = commondao.NewMemberGroup(name, storage)
	urequire.NoError(t, err, "expect no error")
	uassert.Equal(t, name, g.Name(), "expect group name to match")
	uassert.NotNil(t, g.Members(), "expect members to be not nil")
	uassert.Nil(t, g.GetMeta(), "expect default group meta to be nil")
}

func TestMemberGroupMeta(t *testing.T) {
	g, err := commondao.NewMemberGroup("Test", commondao.NewMemberStorage())
	urequire.NoError(t, err, "expect no error")

	g.SetMeta(42)
	v := g.GetMeta()
	urequire.NotEqual(t, nil, v, "expect metadata to be not nil")

	meta, ok := v.(int)
	urequire.True(t, ok, "expect meta type to be int")
	uassert.Equal(t, 42, meta, "expect metadata to match")
}
