package gnolang

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

// R2a's marker: two crossing calls never compare equal, a copy for saving
// drops it, and it fits ObjectInfo's padding (152 bytes; sizes feed gas).
func TestCrossingCallMarker(t *testing.T) {
	a, b := &crossingCall{}, &crossingCall{}
	require.True(t, a != b)
	oi := ObjectInfo{}
	oi.setMadeIn(a)
	require.True(t, oi.madeInCall() == a)
	cp := oi.Copy()
	require.Nil(t, cp.madeInCall())
	require.Equal(t, uintptr(152), unsafe.Sizeof(ObjectInfo{}))
}
