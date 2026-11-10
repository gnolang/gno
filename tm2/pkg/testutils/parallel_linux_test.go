package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMemInfo(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		in   string
		want MemInfo
		ok   bool
	}{
		{
			name: "both present",
			in:   "MemTotal:       32950272 kB\nMemFree:         1048576 kB\nMemAvailable:   20971520 kB\n",
			// kB, not bytes: a mix-up would be off by 1024 either way.
			want: MemInfo{Total: 32950272 << 10, Available: 20 << 30},
			ok:   true,
		},
		{
			// A reading, not a failure: a machine with nothing to spare is
			// when a caller most needs to know.
			name: "nothing available",
			in:   "MemTotal:       32950272 kB\nMemAvailable:          0 kB\n",
			want: MemInfo{Total: 32950272 << 10},
			ok:   true,
		},
		{
			// Kernels before 3.14 have no MemAvailable.
			name: "no MemAvailable",
			in:   "MemTotal:       32950272 kB\nMemFree:         1048576 kB\n",
			ok:   false,
		},
		{
			name: "unparsable value",
			in:   "MemTotal:       32950272 kB\nMemAvailable:   lots kB\n",
			ok:   false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseMemInfo(tt.in)
			require.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

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

func TestParseProcCgroup(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name, in string
		path     string
		v1memory bool
	}{
		{name: "v2 only", in: "0::/user.slice/app.scope\n", path: "/user.slice/app.scope"},
		{name: "v2 namespace root", in: "0::/\n", path: "/"},
		{
			// A stray v1 hierarchy without memory leaves v2 in charge of it.
			name: "v2 beside a v1 net_cls",
			in:   "1:net_cls:/\n0::/user.slice/app.scope\n",
			path: "/user.slice/app.scope",
		},
		{
			name:     "v1 memory",
			in:       "5:memory:/docker/abc\n1:name=systemd:/docker/abc\n0::/\n",
			path:     "/",
			v1memory: true,
		},
		{name: "v1 memory sharing a hierarchy", in: "4:cpu,memory:/x\n", v1memory: true},
		{name: "v1 named hierarchy is not memory", in: "1:name=memory-ish:/x\n"},
		{name: "empty", in: ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path, v1memory := parseProcCgroup(tt.in)
			assert.Equal(t, tt.path, path)
			assert.Equal(t, tt.v1memory, v1memory)
		})
	}
}

func TestCgroupDirs(t *testing.T) {
	t.Parallel()

	mount := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(mount, "a", "b", "c"), 0o700))

	for _, tt := range []struct {
		name, rel string
		want      []string
	}{
		{name: "leaf and every ancestor", rel: "/a/b/c", want: []string{"a/b/c", "a/b", "a", ""}},
		{name: "namespace root", rel: "/", want: []string{""}},
		// A host path seen from inside a container sharing the namespace.
		{name: "missing", rel: "/system.slice/docker-1.scope", want: []string{""}},
		// Paths outside a cgroup namespace show as "/../..".
		{name: "escapes the mount", rel: "/../../a", want: []string{""}},
		{name: "no v2 entry", rel: "", want: nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var want []string
			for _, w := range tt.want {
				want = append(want, filepath.Join(mount, w))
			}
			assert.Equal(t, want, cgroupDirs(mount, tt.rel))
		})
	}
}

func TestCgroupMemory(t *testing.T) {
	t.Parallel()

	const (
		gib4  = "4294967296\n"
		gib1  = "1073741824\n"
		stat  = "anon 1048576\nfile 3221225472\nactive_file 1073741824\ninactive_file 2147483648\n"
		limit = 4 << 30
	)
	for _, tt := range []struct {
		name     string
		files    map[string]string
		wantUsed uint64
		wantOK   bool
	}{
		{
			// At its limit, mostly in cache: the inactive cache is not usage.
			name:     "cache at the limit",
			files:    map[string]string{"memory.max": gib4, "memory.current": gib4, "memory.stat": stat},
			wantUsed: 2 << 30, wantOK: true,
		},
		{
			name:     "no memory.stat",
			files:    map[string]string{"memory.max": gib4, "memory.current": gib1},
			wantUsed: 1 << 30, wantOK: true,
		},
		{
			// The files are not read atomically; must not underflow.
			name:     "more inactive than current",
			files:    map[string]string{"memory.max": gib4, "memory.current": gib1, "memory.stat": stat},
			wantUsed: 0, wantOK: true,
		},
		{name: "no limit at this level", files: map[string]string{"memory.max": "max\n", "memory.current": gib1}},
		{name: "no memory controller"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			for name, content := range tt.files {
				require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
			}
			gotLimit, gotUsed, ok := cgroupMemory(dir)
			require.Equal(t, tt.wantOK, ok)
			if ok {
				assert.Equal(t, uint64(limit), gotLimit)
				assert.Equal(t, tt.wantUsed, gotUsed)
			}
		})
	}
}
