package params

import (
	"testing"
)

// These tests cover proposal CONSTRUCTION. Reading a value back is possible --
// sys/params exposes typed getters (GetSysParamStrings and friends), which
// valset.gno and delegate_test.gno both use, and tests can seed with
// testing.SetSysParam* -- so add read-back assertions rather than assuming they
// cannot be written. End-to-end execution is additionally covered by the
// propX_filetest.gno files under r/gov/dao/.

func TestNewStringPropRequest(cur realm, t *testing.T) {
	pr := NewSysParamStringPropRequest(cur, "foo", "bar", "baz", "qux")
	if pr.Title() == "" {
		t.Errorf("executor shouldn't be nil")
	}
}

func TestNewSetHaltRequest(cur realm, t *testing.T) {
	pr := NewSetHaltRequest(cur, 100_000, "chain/gnoland1.1")
	if pr.Title() == "" {
		t.Errorf("proposal title shouldn't be empty")
	}
}

func TestNewSetHaltRequestNoVersion(cur realm, t *testing.T) {
	pr := NewSetHaltRequest(cur, 100_000, "")
	if pr.Title() == "" {
		t.Errorf("proposal title shouldn't be empty")
	}
}

func TestNewSetHaltRequestCancel(cur realm, t *testing.T) {
	pr := NewSetHaltRequest(cur, 0, "")
	if pr.Title() == "" {
		t.Errorf("proposal title shouldn't be empty")
	}
}
