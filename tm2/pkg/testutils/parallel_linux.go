package testutils

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// cgroupMount is where cgroup v2 is mounted on every distribution that ships it.
const cgroupMount = "/sys/fs/cgroup"

// ReadMemInfo reports the memory this process may use, from /proc/meminfo
// narrowed by the memory limit of its cgroup. It returns false if neither
// figure could be read.
func ReadMemInfo() (MemInfo, bool) {
	// A few kB of virtual file; read it whole rather than scanning it.
	bz, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return MemInfo{}, false
	}

	var mi MemInfo
	for line := range strings.Lines(string(bz)) {
		key, rest, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		var dst *uint64
		switch key {
		case "MemTotal":
			dst = &mi.Total
		case "MemAvailable":
			dst = &mi.Available
		default:
			continue
		}
		// The value is in kB: "MemTotal:       32950272 kB".
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		if kb, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
			*dst = kb * 1024
		}
	}

	if limit, used, ok := cgroupMemory(); ok {
		mi = clampToCgroup(mi, limit, used)
	}
	return mi, mi.Total > 0 && mi.Available > 0
}

// clampToCgroup narrows a host reading to what a cgroup memory limit allows.
//
// /proc/meminfo is not namespaced, so inside a container it describes the host:
// a caller sizing itself against it would be OOM-killed by the cgroup rather
// than throttled by its own accounting. Total is narrowed as well as Available,
// since callers scale their headroom off the memory they believe they have.
func clampToCgroup(mi MemInfo, limit, used uint64) MemInfo {
	mi.Total = min(mi.Total, limit)
	var free uint64
	if limit > used {
		free = limit - used
	}
	mi.Available = min(mi.Available, free)
	return mi
}

// cgroupMemory reports this process's cgroup v2 memory limit and current usage.
// It returns false when there is no v2 limit in force — cgroup v1, or the
// common case of "max". A memory.high throttle, if one is set below the limit,
// is not considered.
func cgroupMemory() (limit, used uint64, ok bool) {
	dir, ok := cgroupDir()
	if !ok {
		return 0, 0, false
	}
	if limit, ok = readCgroupUint(filepath.Join(dir, "memory.max")); !ok {
		return 0, 0, false
	}
	if used, ok = readCgroupUint(filepath.Join(dir, "memory.current")); !ok {
		return 0, 0, false
	}
	return limit, used, true
}

// cgroupDir locates the directory holding this process's cgroup v2 files.
func cgroupDir() (string, bool) {
	bz, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return "", false
	}
	for line := range strings.Lines(string(bz)) {
		// v2 has a single unified entry, "0::<path>"; v1 lines name a
		// controller instead and are skipped.
		rel, ok := strings.CutPrefix(strings.TrimSpace(line), "0::")
		if !ok {
			continue
		}
		// The path is relative to the cgroup root this process can see, which
		// under a cgroup namespace is already the mount. Where it is not — a
		// container sharing the host's namespace — the path names a directory
		// that does not exist here, and the mount root is the right answer.
		if dir := filepath.Join(cgroupMount, rel); fileExists(filepath.Join(dir, "memory.max")) {
			return dir, true
		}
		return cgroupMount, fileExists(filepath.Join(cgroupMount, "memory.max"))
	}
	return "", false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readCgroupUint reads a single-value cgroup file. The literal "max" means no
// limit is in force, which is reported as absent rather than as a huge number.
func readCgroupUint(path string) (uint64, bool) {
	bz, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	s := strings.TrimSpace(string(bz))
	if s == "max" {
		return 0, false
	}
	v, err := strconv.ParseUint(s, 10, 64)
	return v, err == nil
}
