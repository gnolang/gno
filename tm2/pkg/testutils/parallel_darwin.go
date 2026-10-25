package testutils

import (
	"errors"

	"golang.org/x/sys/unix"
)

// ReadMemInfo reports the machine's memory from sysctl. Available is free pages
// plus file-backed ones, the cache the kernel can drop without writing anything
// out: the counterpart of Linux's MemAvailable, and the figure Chromium's
// base::SysInfo uses on macOS.
//
// kern.memorystatus_level looks like the obvious source but is not: on macOS
// it counts active and inactive pages as available, so it only falls once wired
// memory or the compressor grows, long after the heap has.
func ReadMemInfo() (MemInfo, bool) {
	total, err1 := unix.SysctlUint64("hw.memsize")
	// The page counters are in kernel pages, of vm.pagesize bytes: 16 KiB on
	// Apple silicon, 4 KiB on Intel.
	pageSize, err2 := unix.SysctlUint32("vm.pagesize")
	free, err3 := unix.SysctlUint32("vm.page_free_count")
	fileBacked, err4 := unix.SysctlUint32("vm.page_pageable_external_count")
	if errors.Join(err1, err2, err3, err4) != nil {
		return MemInfo{}, false
	}
	mi := MemInfo{
		Total:     total,
		Available: uint64(pageSize) * (uint64(free) + uint64(fileBacked)),
	}
	// Nothing available is a reading like any other.
	return mi, mi.Total > 0
}
