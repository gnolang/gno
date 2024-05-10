package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/gnolang/gno/benchmarking"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

const recordSize int = 10

var pathFlag = flag.String("path", "", "the path to the benchmark file")

func main() {
	flag.Parse()

	file, err := os.Open(*pathFlag)
	if err != nil {
		panic("could not create benchmark file: " + err.Error())
	}
	defer file.Close()

	inputCh := make(chan []byte, 10000)
	outputCh := make(chan string, 10000)
	wg := sync.WaitGroup{}
	numWorkers := 1
	wg.Add(numWorkers)

	doneCh := make(chan struct{})

	for i := 0; i < numWorkers; i++ {
		go func() {
			for {
				record, ok := <-inputCh
				if !ok {
					break
				}

				opName := gnolang.Op(record[0]).String()
				if record[1] != 0 {
					opName = benchmarking.StoreCodeString(record[1])
				}

				elapsedTime := binary.LittleEndian.Uint32(record[2:])
				size := binary.LittleEndian.Uint32(record[6:])
				outputCh <- opName + "," + fmt.Sprint(elapsedTime) + "," + fmt.Sprint(size)
			}
			wg.Done()
		}()
	}

	go func() {
		out, err := os.Create("results.csv")
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

			fmt.Fprintln(out, output)
		}

		out.Close()
		doneCh <- struct{}{}
	}()

	var i int

	bufSize := recordSize * 100000
	buf := make([]byte, bufSize)

	for {
		nbytes, err := file.Read(buf)

		if err != nil && nbytes == 0 {
			break
		}
		n := nbytes / recordSize

		for j := 0; j < n; j++ {
			inputCh <- buf[j*recordSize : (j+1)*recordSize]
		}

		i += bufSize / recordSize
		if i%1000 == 0 {
			fmt.Println(i)
		}
	}

	close(inputCh)
	wg.Wait()
	close(outputCh)
	<-doneCh
	close(doneCh)

	fmt.Println("done")
}
