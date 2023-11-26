package postit

import (
	"strconv"
	"testing"
)

func TestUpdateTrending(t *testing.T) {
	for i := uint64(0); i < 50; i++ {
		updateTrending(strconv.FormatUint(i, 10), i)
	}
	for i := uint64(0); i < 50; i++ {
		updateTrending(strconv.FormatUint(i, 10), i+1)
	}

	tmp := uint64(49)
	for _, tr := range trending {
		tag := strconv.FormatUint(tmp, 10)
		count := tmp + 1
		if tr.Tag != tag {
			t.Errorf("tag: expected: %s; got %s", tag, tr.Tag)
		}
		if tr.Count != count {
			t.Errorf("count: expected: %d; got %d", count, tr.Count)
		}
		tmp--
	}
}
