package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"sync"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type codeStats struct {
	codeName   string
	avgTime    int64
	avgSize    int64
	timeStdDev int64
	count      int
}

type codeRecord struct {
	codeName string
	elapsed  uint32
	size     uint32
}

// It reads binary record, calcuate and output the statistics of operations
func stats(binFile string) {
	in, err := os.Open(binFile)
	if err != nil {
		panic("could not create benchmark file: " + err.Error())
	}
	defer in.Close()

	inputCh := make(chan []byte, 10000)
	outputCh := make(chan codeRecord, 10000)
	wg := sync.WaitGroup{}
	numWorkers := 2
	wg.Add(numWorkers)
	doneCh := make(chan struct{})
	for range numWorkers {
		go func() {
			for {
				record, ok := <-inputCh
				if !ok {
					break
				}
				var opName string
				switch record[0] {
				case byte(bm.TypeOpCode):
					opName = gno.Op(record[1]).String()
				case byte(bm.TypeStore):
					opName = bm.StoreCodeString(record[1])
				case byte(bm.TypeNative):
					opName = bm.NativeCodeString(record[1])
				default:
					panic(fmt.Sprintf("invalid record type: %d", record[0]))
				}

				elapsedTime := binary.LittleEndian.Uint32(record[2:])
				size := binary.LittleEndian.Uint32(record[6:])
				outputCh <- codeRecord{opName, elapsedTime, size}
			}
			wg.Done()
		}()
	}

	crs := []codeRecord{}
	// out put
	go func() {
		out, err := os.Create(*outFlag)
		if err != nil {
			panic("could not create readable output file: " + err.Error())
		}
		defer out.Close()
		fmt.Fprintln(out, "op,elapsedTime,diskIOBytes")

		for {
			output, ok := <-outputCh
			if !ok {
				break
			}
			csv := output.codeName + "," + fmt.Sprint(output.elapsed) + "," + fmt.Sprint(output.size)
			fmt.Fprintln(out, csv)
			crs = append(crs, output)
		}

		out.Close()
		doneCh <- struct{}{}
	}()

	recordSize := bm.RecordSize
	bufSize := recordSize * 100000
	buf := make([]byte, bufSize)

	for {
		nbytes, err := in.Read(buf)

		if err != nil && nbytes == 0 {
			break
		}
		n := nbytes / recordSize

		for j := range n {
			inputCh <- buf[j*recordSize : (j+1)*recordSize]
		}
	}

	close(inputCh)
	wg.Wait()
	close(outputCh)
	<-doneCh
	close(doneCh)

	calculateStats(crs)
	fmt.Println("done")
}

func calculateStats(crs []codeRecord) {
	filename := *outFlag
	out, err := os.Create(addSuffix(filename))
	if err != nil {
		panic("could not create readable output file: " + err.Error())
	}
	defer out.Close()
	fmt.Fprintln(out, "op,avg_time,avg_size,time_stddev,count")

	m := make(map[string][]codeRecord)
	for _, v := range crs {
		crs, ok := m[v.codeName]
		if ok {
			crs = append(crs, v)
			m[v.codeName] = crs
		} else {
			m[v.codeName] = []codeRecord{v}
		}
	}

	keys := make([]string, 0, 100)

	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		cs := calculate(k, m[k])
		csv := cs.codeName + "," + fmt.Sprint(cs.avgTime) + "," + fmt.Sprint(cs.avgSize) + "," + fmt.Sprint(cs.timeStdDev) + "," + fmt.Sprint(cs.count)
		fmt.Fprintln(out, csv)
	}

	fmt.Println("## Benchmark results saved in:", filename)
	fmt.Println("## Benchmark result stats saved in:", out.Name())
}

func addSuffix(filename string) string {
	// Find the position of the last dot
	dotPos := strings.LastIndex(filename, ".")
	if dotPos == -1 {
		// No dot found, return the original filename with '_status' appended
		return filename + "_stats"
	}
	// Insert '_status' before the last suffix
	return filename[:dotPos] + "_stats" + filename[dotPos:]
}

// calcuate the average and standard deviation in  time of a code name

func calculate(codeName string, crs []codeRecord) codeStats {
	// Calculate average
	var sumTime int64
	var sumSize int64
	for _, cr := range crs {
		t := cr.elapsed
		s := cr.size
		sumTime += int64(t)
		sumSize += int64(s)
	}
	avgTime := float64(sumTime) / float64(len(crs))
	avgSize := float64(sumSize) / float64(len(crs))

	// Calculate standard deviation of duration in time
	var varianceSum float64
	for _, cr := range crs {
		varianceSum += math.Pow(float64(cr.elapsed)-avgTime, 2)
	}
	variance := varianceSum / float64(len(crs))
	stdDev := math.Sqrt(variance)
	return codeStats{codeName, int64(avgTime), int64(avgSize), int64(stdDev), len(crs)}
}
