// Mirrors of stdlibs/testing Report / BenchmarkReport / BenchmarkReports.
// Wire format must stay in lock-step with testing.gno's marshal() methods.

package test

type report struct {
	Failed  bool
	Skipped bool
}

type benchmarkReport struct {
	report

	Name         string
	ReportAllocs bool
	N            int
	Cycles       int64
	Gas          int64
	AllocBytes   int64
	Allocs       int64
	Bytes        int64
}

type benchmarkReports struct {
	Reports []benchmarkReport
}
