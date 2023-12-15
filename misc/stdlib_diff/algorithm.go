package main

type Algorithm interface {
	Diff() (srcDiff []LineDifferrence, dstDiff []LineDifferrence)
}
