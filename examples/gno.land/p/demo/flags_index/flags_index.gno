package flags_index

import (
	"strconv"

	"gno.land/p/demo/cpavl"
)

type FlagID string

type FlagCount struct {
	FlagID FlagID
	Count  uint64
}

type FlagsIndex struct {
	flagsCounts      []*FlagCount // sorted by count descending; TODO: optimize using big brain datastructure
	flagsCountsByID  *avl.Tree    // key: flagID -> FlagCount
	flagsByFlaggerID *avl.Tree    // key: flaggerID -> *avl.Tree key: flagID -> struct{}
}

func NewFlagsIndex() *FlagsIndex {
	return &FlagsIndex{
		flagsCountsByID:  avl.NewTree(),
		flagsByFlaggerID: avl.NewTree(),
	}
}

func (fi *FlagsIndex) HasFlagged(flagID FlagID, flaggerID string) bool {
	if flagsByFlagID, ok := fi.flagsByFlaggerID.Get(flaggerID); ok {
		if flagsByFlagID.(*avl.Tree).Has(string(flagID)) {
			return true
		}
	}
	return false
}

func (fi *FlagsIndex) GetFlagCount(flagID FlagID) uint64 {
	if flagCount, ok := fi.flagsCountsByID.Get(string(flagID)); ok {
		return flagCount.(*FlagCount).Count
	}
	return 0
}

func (fi *FlagsIndex) GetFlags(limit uint64, offset uint64) []*FlagCount {
	if limit == 0 {
		return nil
	}
	if offset >= uint64(len(fi.flagsCounts)) {
		return nil
	}
	if offset+limit > uint64(len(fi.flagsCounts)) {
		limit = uint64(len(fi.flagsCounts)) - offset
	}
	return fi.flagsCounts[offset : offset+limit]
}

func (fi *FlagsIndex) Flag(flagID FlagID, flaggerID string) {
	// update flagsByFlaggerID
	var flagsByFlagID *avl.Tree
	if existingFlagsByFlagID, ok := fi.flagsByFlaggerID.Get(flaggerID); ok {
		flagsByFlagID = existingFlagsByFlagID.(*avl.Tree)
		if flagsByFlagID.(*avl.Tree).Has(string(flagID)) {
			panic("already flagged")
		}
	} else {
		newFlagsByFlagID := avl.NewTree()
		fi.flagsByFlaggerID.Set(flaggerID, newFlagsByFlagID)
		flagsByFlagID = newFlagsByFlagID
	}
	flagsByFlagID.Set(string(flagID), struct{}{})

	// update flagsCountsByID and flagsCounts
	iFlagCount, ok := fi.flagsCountsByID.Get(string(flagID))
	if !ok {
		flagCount := &FlagCount{FlagID: flagID, Count: 1}
		fi.flagsCountsByID.Set(string(flagID), flagCount)
		fi.flagsCounts = append(fi.flagsCounts, flagCount) // this is valid because 1 will always be the lowest count and we want the newest flags to be last
	} else {
		flagCount := iFlagCount.(*FlagCount)
		flagCount.Count++
		// move flagCount to correct position in flagsCounts
		for i := len(fi.flagsCounts) - 1; i > 0; i-- {
			if fi.flagsCounts[i].Count > fi.flagsCounts[i-1].Count {
				fi.flagsCounts[i], fi.flagsCounts[i-1] = fi.flagsCounts[i-1], fi.flagsCounts[i]
			} else {
				break
			}
		}
	}
}

func (fi *FlagsIndex) ClearFlagCount(flagID FlagID) {
	// find flagCount in byID
	if !fi.flagsCountsByID.Has(string(flagID)) {
		panic("flag ID not found")
	}

	// remove from byID
	fi.flagsCountsByID.Remove(string(flagID))

	// remove from byCount, we need to recreate the slice since splicing is broken
	newByCount := make([]*FlagCount, len(fi.flagsCounts)-1)
	for i := range fi.flagsCounts {
		if fi.flagsCounts[i].FlagID == flagID {
			continue
		}
		newByCount = append(newByCount, fi.flagsCounts[i])
	}
	fi.flagsCounts = newByCount

	// update flagsByFlaggerID
	var empty []string
	fi.flagsByFlaggerID.Iterate("", "", func(key string, value interface{}) bool {
		t := value.(*avl.Tree)
		t.Remove(string(flagID))
		if t.Size() == 0 {
			empty = append(empty, key)
		}
		return false
	})
	for _, key := range empty {
		fi.flagsByFlaggerID.Remove(key)
	}
}

func (fi *FlagsIndex) Dump() string {
	str := ""

	str += "## flagsCounts:\n"
	for i := range fi.flagsCounts {
		str += "- " + string(fi.flagsCounts[i].FlagID) + " " + strconv.FormatUint(fi.flagsCounts[i].Count, 10) + "\n"
	}

	str += "\n## flagsCountsByID:\n"
	fi.flagsCountsByID.Iterate("", "", func(key string, value interface{}) bool {
		str += "- " + key + ": " + string(value.(*FlagCount).FlagID) + " " + strconv.FormatUint(value.(*FlagCount).Count, 10) + "\n"
		return false
	})

	str += "\n## flagsByFlaggerID:\n"
	fi.flagsByFlaggerID.Iterate("", "", func(key string, value interface{}) bool {
		str += "- " + key + ":\n"
		value.(*avl.Tree).Iterate("", "", func(key string, value interface{}) bool {
			str += "  - " + key + "\n"
			return false
		})
		return false
	})

	return str
}
