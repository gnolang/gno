package testutils

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// cgroupMount is where cgroup v2 is mounted on every distribution that ships it.
const cgroupMount = "/sys/fs/cgroup"

// ReadMemInfo reports the memory this process may use, from /proc/meminfo
// narrowed by the memory limits of its cgroup and of every cgroup above it. It
// returns false if neither figure could be read, or if memory is under a
// cgroup v1 controller: its limits are not read here, and the host figures can
// sit far beyond them.
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

	// Unreadable, this parses as no cgroup at all, and the host figures stand.
	cg, _ := os.ReadFile("/proc/self/cgroup")
	rel, v1memory := parseProcCgroup(string(cg))
	if v1memory {
		return MemInfo{}, false
	}
	for _, dir := range cgroupDirs(cgroupMount, rel) {
		if limit, used, ok := cgroupMemory(dir); ok {
			mi = clampToCgroup(mi, limit, used)
		}
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

// cgroupMemory reports one cgroup v2 directory's memory limit and usage. It
// returns false when no limit is in force there — the common case of "max", or
// no memory controller at that level. A memory.high throttle, if one is set
// below the limit, is not considered. Usage leaves out inactive_file, as
// kubelet does: page cache the kernel reclaims first, which would otherwise
// hold a cgroup that has read files at its limit for good.
func cgroupMemory(dir string) (limit, used uint64, ok bool) {
	if limit, ok = readCgroupUint(filepath.Join(dir, "memory.max")); !ok {
		return 0, 0, false
	}
	if used, ok = readCgroupUint(filepath.Join(dir, "memory.current")); !ok {
		return 0, 0, false
	}
	if inactive, ok := readCgroupStat(filepath.Join(dir, "memory.stat"), "inactive_file"); ok {
		used -= min(used, inactive)
	}
	return limit, used, true
}

// parseProcCgroup reads the contents of /proc/self/cgroup, whose lines are
// "<hierarchy-ID>:<controllers>:<path>". It reports the cgroup v2 path — the
// single entry with ID 0 and no controllers, empty if there is none — and
// whether a v1 hierarchy holds the memory controller.
func parseProcCgroup(s string) (v2path string, v1memory bool) {
	for line := range strings.Lines(s) {
		id, rest, _ := strings.Cut(strings.TrimSpace(line), ":")
		controllers, path, ok := strings.Cut(rest, ":")
		if !ok {
			continue
		}
		if id == "0" && controllers == "" {
			v2path = path
		} else if slices.Contains(strings.Split(controllers, ","), "memory") {
			v1memory = true
		}
	}
	return v2path, v1memory
}

// cgroupDirs lists the cgroup v2 directories whose memory limits bind a
// process at rel: its own, then each ancestor up to the mount. A limit may sit
// on any of them — a systemd slice, say, above a scope left at "max".
//
// The path is relative to the cgroup root this process can see, which under a
// cgroup namespace is already the mount. Where it is not — a container sharing
// the host's namespace — the path names a directory that does not exist here,
// and only the mount itself applies.
func cgroupDirs(mount, rel string) []string {
	if rel == "" {
		return nil
	}
	dir := filepath.Join(mount, rel)
	if _, err := os.Stat(dir); err != nil || !strings.HasPrefix(dir, mount+string(filepath.Separator)) {
		dir = mount
	}
	dirs := []string{dir}
	for dir != mount {
		dir = filepath.Dir(dir)
		dirs = append(dirs, dir)
	}
	return dirs
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

// readCgroupStat reads one key from a flat-keyed cgroup file such as
// memory.stat, whose lines are "<key> <value>".
func readCgroupStat(path, key string) (uint64, bool) {
	bz, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	for line := range strings.Lines(string(bz)) {
		if k, v, ok := strings.Cut(strings.TrimSpace(line), " "); ok && k == key {
			n, err := strconv.ParseUint(v, 10, 64)
			return n, err == nil
		}
	}
	return 0, false
}
