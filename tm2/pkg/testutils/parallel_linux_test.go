package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClampToCgroup(t *testing.T) {
	t.Parallel()

	const host = 32 << 30
	for _, tt := range []struct {
		name              string
		mi                MemInfo
		limit, used       uint64
		wantTotal, wantAv uint64
	}{
		{
			// The case this exists for: a container far smaller than its host.
			name: "limit below host",
			mi:   MemInfo{Total: host, Available: 20 << 30},
			// A quarter of 4 GiB is a very different reserve from a quarter of
			// 32, so Total has to be narrowed too, not just Available.
			limit: 4 << 30, used: 1 << 30,
			wantTotal: 4 << 30, wantAv: 3 << 30,
		},
		{
			name: "host tighter than the limit",
			mi:   MemInfo{Total: host, Available: 2 << 30},
			// The host is already short of memory; the cgroup's generous
			// headroom does not conjure any.
			limit: 16 << 30, used: 1 << 30,
			wantTotal: 16 << 30, wantAv: 2 << 30,
		},
		{
			name:  "usage at the limit",
			mi:    MemInfo{Total: host, Available: 20 << 30},
			limit: 4 << 30, used: 4 << 30,
			wantTotal: 4 << 30, wantAv: 0,
		},
		{
			// memory.current can exceed memory.max transiently, before reclaim
			// catches up. Must not underflow into a huge Available.
			name:  "usage past the limit",
			mi:    MemInfo{Total: host, Available: 20 << 30},
			limit: 4 << 30, used: 5 << 30,
			wantTotal: 4 << 30, wantAv: 0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := clampToCgroup(tt.mi, tt.limit, tt.used)
			assert.Equal(t, tt.wantTotal, got.Total, "Total")
			assert.Equal(t, tt.wantAv, got.Available, "Available")
			assert.LessOrEqual(t, got.Available, got.Total)
		})
	}
}

func TestReadCgroupUint(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		return path
	}

	v, ok := readCgroupUint(write("bytes", "4294967296\n"))
	assert.True(t, ok)
	assert.Equal(t, uint64(4<<30), v)

	// "max" means no limit in force, and must not read as a number.
	_, ok = readCgroupUint(write("unlimited", "max\n"))
	assert.False(t, ok)

	_, ok = readCgroupUint(write("garbage", "not a number"))
	assert.False(t, ok)

	_, ok = readCgroupUint(dir + "/does-not-exist")
	assert.False(t, ok)
}
