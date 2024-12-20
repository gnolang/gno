package btree_dao

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/btree"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func setupTest() {
	std.TestSetOrigCaller(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))
	members = btree.New()
}

type TestElement struct {
	value int
}

func (te *TestElement) Less(than btree.Record) bool {
	return te.value < than.(*TestElement).value
}

func TestPlantTree(t *testing.T) {
	setupTest()

	tree := btree.New()
	elements := []int{30, 10, 50, 20, 40}
	for _, val := range elements {
		tree.Insert(&TestElement{value: val})
	}

	err := PlantTree(tree)
	urequire.NoError(t, err, "Should successfully plant a valid tree")

	found := false
	members.Ascend(func(record btree.Record) bool {
		regDetails := record.(*RegistrationDetails)
		if regDetails.UserBTree == tree {
			found = true
			return false
		}
		return true
	})
	uassert.True(t, found, "Tree should be registered in members")

	err = PlantTree(tree)
	uassert.Error(t, err, "Should not allow planting the same tree twice")

	emptyTree := btree.New()
	err = PlantTree(emptyTree)
	uassert.Error(t, err, "Should not allow planting an empty tree")
}

func TestPlantSeed(t *testing.T) {
	setupTest()

	err := PlantSeed("Hello DAO!")
	urequire.NoError(t, err, "Should successfully plant a seed with valid message")

	found := false
	members.Ascend(func(record btree.Record) bool {
		regDetails := record.(*RegistrationDetails)
		if regDetails.UserBTree == nil {
			found = true
			return false
		}
		return true
	})
	uassert.True(t, found, "Seed should be registered in members")

	err = PlantSeed("")
	uassert.Error(t, err, "Should not allow planting seed with empty message")
}

func TestRegistrationDetailsOrdering(t *testing.T) {
	setupTest()

	rd1 := &RegistrationDetails{
		Address: std.Address("test1"),
		RegTime: time.Now(),
	}
	rd2 := &RegistrationDetails{
		Address: std.Address("test2"),
		RegTime: time.Now().Add(time.Hour),
	}

	uassert.True(t, rd1.Less(rd2), "Earlier registration should be less than later one")
	uassert.False(t, rd2.Less(rd1), "Later registration should not be less than earlier one")
}
